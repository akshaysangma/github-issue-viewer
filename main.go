package main

import (
	"log"
	"net/http"
	"os"

	"github.com/akshaysangma/github-issue-viewer/api"
	"github.com/akshaysangma/github-issue-viewer/graphql"
	"github.com/akshaysangma/github-issue-viewer/oauth"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)

	err := godotenv.Load()

	if err != nil {
		logger.Fatal("Error loading .env file")
	}

	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	redirectURI := os.Getenv("REDIRECT_URI")
	authEndpoint := os.Getenv("AUTH_ENDPOINT")
	tokenEndpoint := os.Getenv("TOKEN_ENDPOINT")
	graphqlEndpoint := os.Getenv("GRAPHQL_ENDPOINT")

	oAuth, err := oauth.New(oauth.WithClientID(clientID), oauth.WithClientSecret(clientSecret), oauth.WithRedirectURI(redirectURI), oauth.WithEndpointURL(authEndpoint, tokenEndpoint))

	gqlCfg := graphql.New(graphqlEndpoint)

	if err != nil {
		logger.WithFields(logrus.Fields{
			"event": "init",
			"error": err.Error(),
		}).Fatal("Error in initializing OAuth")
	}

	svrCfg, err := api.New(api.WithListenAddr(":8080"), api.WithOAuth(oAuth), api.WithLogger(logger), api.WithGraphqlConfig(gqlCfg))

	if err != nil {
		logger.WithFields(logrus.Fields{
			"event": "init",
			"error": err.Error(),
		}).Fatal("Error in initializing server")
	}

	router := chi.NewRouter()

	// TODO : Make it more restrictive
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	fs := http.FileServer(http.Dir("templates"))
	router.Handle("/templates/*", http.StripPrefix("/templates/", fs))
	router.Get("/login", svrCfg.LoginHandler)
	router.Get("/oauth/callback", svrCfg.OAuthCallbackHandler)
	router.Get("/home", svrCfg.AuthMiddleware(svrCfg.HomeHandler))
	router.Get("/find", svrCfg.AuthMiddleware((svrCfg.SearchForRepositoryHandler)))

	srv := &http.Server{
		Addr:    api.GetListenAddr(svrCfg),
		Handler: router,
	}

	log.Println("Server running on ", api.GetListenAddr(svrCfg))
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}
