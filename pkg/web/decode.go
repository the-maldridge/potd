package web

import (
	"net/http"
	"strconv"
	"os"

	"github.com/flosch/pongo2/v6"

	"github.com/the-maldridge/potd/pkg/password"
)

func (s *Server) uiViewDecodeForm(w http.ResponseWriter, r *http.Request) {
	s.doTemplate(w, r, "views/decode/form.p2", nil)
}

func (s *Server) uiViewDecode(w http.ResponseWriter, r *http.Request) {
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

	s.doTemplate(w, r, "views/decode/password.p2",
		pongo2.Context{"passwords": []string{p.String()}},
	)
}
