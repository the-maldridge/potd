package web

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/flosch/pongo2/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/the-maldridge/authware"
)

//go:embed theme
var efs embed.FS

func New(opts ...Option) (*Server, error) {
	var tfs fs.FS
	tfs, _ = fs.Sub(efs, "theme")
	tsfs, _ := fs.Sub(tfs, "p2")

	basic, err := authware.NewAuth()
	if err != nil {
		slog.Error("Could not initialize middleware", "error", err)
		os.Exit(2)
	}

	s := &Server{
		r: chi.NewRouter(),
		n: &http.Server{
			Addr: ":1323",
		},
		tmpls: pongo2.NewSet("html", pongo2.NewFSLoader(tsfs)),
	}

	s.r.Use(middleware.Heartbeat("/-/alive"))
	sfs, _ := fs.Sub(tfs, "static")
	s.r.Handle("/static/*", http.StripPrefix("/static", http.FileServerFS(sfs)))
	s.r.Get("/login", s.uiViewLoginPage)
	s.r.Post("/login", basic.LoginFormHandler("username", "password", "/ui/"))
	s.r.Get("/logout", basic.LogoutHandler("/"))

	s.r.Route("/ui", func(r chi.Router) {
		r.Use(basic.LoginHandler("/login"))
		r.Get("/", s.uiViewLanding)
		r.Get("/resolve", s.uiViewResolveForm)
		r.Post("/resolve", s.uiViewResolve)
	})

	s.r.Route("/api/resolve", func(r chi.Router) {
		r.Use(basic.BasicHandler)
		r.Get("/{host}/{challenge}", s.apiResolvePassword)
	})

	s.r.Route("/api/escrow", func(r chi.Router) {
		r.Post("/update-token", s.apiUpdateToken)
	})

	s.r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ui/resolve", http.StatusSeeOther)
	})

	for _, o := range opts {
		o(s)
	}

	return s, nil
}

// Serve binds, initializes the mux, and serves forever.
func (s *Server) Serve(cert, key string) error {
	slog.Info("HTTP is starting", "bind", s.n.Addr)
	s.n.Handler = s.r
	return s.n.ListenAndServeTLS(cert, key)
}

// Shutdown requests the underlying server gracefully cease operation.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.n.Shutdown(ctx)
}

func (s *Server) templateErrorHandler(w http.ResponseWriter, err error) {
	fmt.Fprintf(w, "Error while rendering template: %s\n", err)
}

func (s *Server) doTemplate(w http.ResponseWriter, r *http.Request, tmpl string, ctx pongo2.Context) {
	if ctx == nil {
		ctx = pongo2.Context{}
	}

	ctx["user"], _ = r.Context().Value(authware.UserKey{}).(authware.User)

	t, err := s.tmpls.FromCache(tmpl)
	if err != nil {
		s.templateErrorHandler(w, err)
		return
	}
	if err := t.ExecuteWriter(ctx, w); err != nil {
		s.templateErrorHandler(w, err)
	}
}

func (s *Server) uiViewLoginPage(w http.ResponseWriter, r *http.Request) {
	s.doTemplate(w, r, "login.p2", nil)
}

func (s *Server) uiViewLanding(w http.ResponseWriter, r *http.Request) {
	s.doTemplate(w, r, "base.p2", nil)
}
