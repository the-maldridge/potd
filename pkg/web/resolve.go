package web

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/flosch/pongo2/v6"
	"github.com/go-chi/chi/v5"
	"github.com/the-maldridge/authware"
	"gorm.io/gorm"

	"github.com/the-maldridge/potd/pkg/password"
	"github.com/the-maldridge/potd/pkg/types"
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

	eToken, err := gorm.G[types.EscrowedToken](s.d).
		Where(&types.EscrowedToken{Host: r.FormValue("hostname")}).
		First(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("no host found"))
		return
	}

	components := []string{
		r.FormValue("hostname"),
		r.FormValue("challenge"),
		eToken.Token,
	}
	slog.Debug("Resolving from componenents", "components", components)

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
		size = 5
	}

	kindID, err := strconv.Atoi(r.URL.Query().Get("password_type"))
	if err != nil {
		kindID = 2
	}
	kind := password.Mode(kindID)

	eToken, err := gorm.G[types.EscrowedToken](s.d).
		Where(&types.EscrowedToken{Host: host}).
		First(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("no host found"))
		return
	}

	components := []string{
		host,
		challenge,
		eToken.Token,
	}
	slog.Debug("Resolving from componenents", "components", components)

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
