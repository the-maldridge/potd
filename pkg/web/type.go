package web

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/flosch/pongo2/v6"
)

// Server contains the server and its associated machinery.
type Server struct {
	r chi.Router
	n *http.Server

	tokenPath string

	tmpls *pongo2.TemplateSet
}
