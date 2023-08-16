package api

import (
	"net/http"
	"text/template"
	"time"
)

const (
	AccessTokenCookieName = "oauth_token"
)

func (s *server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	authURL, err := s.oAuth.GenerateAuthURL()
	if err != nil {
		http.Error(w, "Undefined", http.StatusInternalServerError)
		return
	}

	tmplData := struct {
		AuthURL string
	}{
		AuthURL: authURL,
	}

	tmplParsed, err := template.ParseFiles("templates/login.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tmplParsed.Execute(w, tmplData)
	if err != nil {
		http.Error(w, "Template rendering error", http.StatusInternalServerError)
		return
	}
}

func (s *server) HomeHandler(w http.ResponseWriter, r *http.Request, token string) {

	tmplParsed, err := template.ParseFiles("templates/app.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tmplParsed.Execute(w, nil)
	if err != nil {
		http.Error(w, "Template rendering error", http.StatusInternalServerError)
		return
	}

}

func (s *server) OAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code", http.StatusBadRequest)
		return
	}

	oAuthResp, err := s.oAuth.Exchange(code)
	if err != nil {
		s.logger.WithError(err).Error("OAuth exchange error")
		http.Error(w, "Exchange error", http.StatusInternalServerError)
		return
	}

	// TODO: Find a better strategy to store the token in the browser
	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookieName,
		Value:    oAuthResp.AccessToken,
		Expires:  time.Now().Add(time.Second * time.Duration(oAuthResp.ExpiresIn)),
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Location", "/home")
	w.WriteHeader(http.StatusFound)
}
