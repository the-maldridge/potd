package web

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/flosch/pongo2/v6"
	"github.com/go-chi/chi/v5"
	"github.com/the-maldridge/authware"

	"github.com/the-maldridge/potd/pkg/password"
)

func (s *Server) uiViewResolveForm(w http.ResponseWriter, r *http.Request) {
	s.doTemplate(w, r, "views/resolve/form.p2", nil)
}

func (s *Server) uiViewResolve(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	size, err := strconv.Atoi(r.FormValue("size"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	kindID, err := strconv.Atoi(r.FormValue("password_type"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	kind := password.Mode(kindID)

	content, err := os.ReadFile(s.tokenPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	components := []string{r.FormValue("hostname"), r.FormValue("challenge"), string(content)}

	p := password.New(components, kind, size)

	s.doTemplate(w, r, "views/resolve/password.p2",
		pongo2.Context{"passwords": []string{p.String()}},
	)
	slog.Info("Password Resolved",
		"host", r.FormValue("hostname"),
		"user", r.Context().Value(authware.UserKey{}).(authware.User).Identity,
	)
}

func (s *Server) apiResolvePassword(w http.ResponseWriter, r *http.Request) {
	host := chi.URLParam(r, "host")
	challenge := chi.URLParam(r, "challenge")

	size, err := strconv.Atoi(r.URL.Query().Get("size"))
	if err != nil {
		size = 24
	}

	kindID, err := strconv.Atoi(r.URL.Query().Get("password_type"))
	if err != nil {
		kindID = 1
	}
	kind := password.Mode(kindID)

	content, err := os.ReadFile(s.tokenPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	components := []string{host, challenge, string(content)}

	p := password.New(components, kind, size)

	res := struct{ Passwords []string }{Passwords: []string{p.String()}}
	if err := json.NewEncoder(w).Encode(res); err != nil {
		slog.Error("Error encoding response", "error", err)
	}
	slog.Info("Password Resolved",
		"host", host,
		"user", r.Context().Value(authware.UserKey{}).(authware.User).Identity,
	)
}
