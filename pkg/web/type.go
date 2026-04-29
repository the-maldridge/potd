package web

import (
	"net/http"

	"github.com/flosch/pongo2/v6"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// Server contains the server and its associated machinery.
type Server struct {
	r chi.Router
	n *http.Server
	d *gorm.DB

	trimPrefix string
	trimSuffix string

	tmpls *pongo2.TemplateSet
}

// An Option is used to configure the server.
type Option func(*Server)
