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

func New() (*Server, error) {
	var tfs fs.FS
	tfs, _ = fs.Sub(efs, "theme")
	if path, ok := os.LookupEnv("POTD_DEBUG"); ok {
		slog.Info("Debug mode enabled")
		tfs = os.DirFS(path)
	}
	tsfs, _ := fs.Sub(tfs, "p2")

	basic, err := authware.NewAuth()
	if err != nil {
		slog.Error("Could not initialize middleware", "error", err)
		os.Exit(2)
	}
	tPath := os.Getenv("POTD_SHARED_TOKEN_PATH")
	if tPath == "" {
		tPath = "/usr/share/potd/shared_token"
	}

	s := &Server{
		r:         chi.NewRouter(),
		n:         &http.Server{},
		tokenPath: tPath,
		tmpls:     pongo2.NewSet("html", pongo2.NewFSLoader(tsfs)),
	}

	if _, ok := os.LookupEnv("POTD_DEBUG"); ok {
		s.tmpls.Debug = true
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
		r.Get("/decode", s.uiViewDecodeForm)
		r.Post("/decode", s.uiViewDecode)
	})

	s.r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ui/decode", http.StatusSeeOther)
	})

	return s, nil
}

// Serve binds, initializes the mux, and serves forever.
func (s *Server) Serve(bind string) error {
	slog.Info("HTTP is starting", "bind", bind)
	s.n.Addr = bind
	s.n.Handler = s.r
	return s.n.ListenAndServe()
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
