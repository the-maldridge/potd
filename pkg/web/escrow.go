package web

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/the-maldridge/potd/pkg/types"
)

func (s *Server) apiUpdateToken(w http.ResponseWriter, r *http.Request) {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		slog.Info("error authenticating client", "peer", r.TLS.PeerCertificates, "tls", r.TLS)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "Certificate was not supplied")
		return
	}

	cn := r.TLS.PeerCertificates[0].Subject.CommonName
	host := strings.Split(cn, ".")[0]
	slog.Info("Inbound request", "cn", cn, "host", host)

	et := types.EscrowedToken{}
	if err := json.NewDecoder(r.Body).Decode(&et); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error in request: %s\n", err)
		return
	}
	et.Updated = time.Now()

	if s.trimPrefix != "" || s.trimSuffix != "" {
		slog.Debug("Trimming host identity", "prefix", s.trimPrefix, "suffix", s.trimSuffix, "host", host)
		host = strings.TrimPrefix(strings.TrimSuffix(host, s.trimSuffix), s.trimPrefix)
	}
	if host != et.Host {
		w.WriteHeader(http.StatusUnauthorized)
		slog.Warn("Host does not match identity", "host", et.Host, "identity", host)
		fmt.Fprintf(w, "Host does not match identity: %s != %s\n", host, et.Host)
		return
	}

	err := gorm.G[types.EscrowedToken](s.d, clause.OnConflict{
		UpdateAll: true,
	}).Create(r.Context(), &et)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error saving token: %s\n", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	slog.Info("Updated escrow token for host", "host", et.Host)
}
