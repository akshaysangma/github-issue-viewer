package api

import (
	"net/http"
	"text/template"
	"time"
)

type authedHandler func(http.ResponseWriter, *http.Request, string)

func (s *server) AuthMiddleware(Handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(AccessTokenCookieName)
		if err != nil {
			s.Relog(w, r)
			return
		}
		if isTokenExpired := c.Expires.After(time.Now()); isTokenExpired {
			s.Relog(w, r)
			return
		}
		Handler(w, r, c.Value)
	}
}

func (s *server) Relog(w http.ResponseWriter, r *http.Request) {
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
