package web

import (
	"crypto/tls"
	"crypto/x509"
	"gorm.io/gorm"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/flosch/pongo2/v6"
)

func WithTrimPrefix(prefix string) Option {
	return func(s *Server) {
		s.trimPrefix = prefix
	}
}

func WithTrimSuffix(suffix string) Option {
	return func(s *Server) {
		s.trimSuffix = suffix
	}
}

func WithBind(bind string) Option {
	return func(s *Server) {
		s.n.Addr = bind
	}
}

func WithDB(d *gorm.DB) Option {
	return func(s *Server) {
		s.d = d
	}
}

func WithTemplatePath(p string) Option {
	return func(s *Server) {
		slog.Debug("Configuring debug template path", "path", p)
		tfs := os.DirFS(p)
		tsfs, _ := fs.Sub(tfs, "p2")
		s.tmpls = pongo2.NewSet("html", pongo2.NewFSLoader(tsfs))
		s.tmpls.Debug = true
		sfs, _ := fs.Sub(tfs, "static")
		s.r.Handle("/static/*", http.StripPrefix("/static", http.FileServerFS(sfs)))
	}
}

func WithClientCA(c string) Option {
	return func(s *Server) {
		cf, err := os.ReadFile(c)
		if err != nil {
			slog.Error("Could not read CA file", "error", err)
			return
		}
		p := x509.NewCertPool()
		p.AppendCertsFromPEM(cf)

		tc := tls.Config{
			ClientAuth: tls.VerifyClientCertIfGiven,
			ClientCAs:  p,
		}
		s.n.TLSConfig = &tc
	}
}
