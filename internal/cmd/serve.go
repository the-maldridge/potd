package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	_ "github.com/the-maldridge/authware/backend/htpasswd"
	_ "github.com/the-maldridge/authware/backend/ldap"
	_ "github.com/the-maldridge/authware/backend/netauth"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/the-maldridge/potd/pkg/types"
	"github.com/the-maldridge/potd/pkg/web"
)

var (
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "serve a decoding server",
		Run:   serveCmdRun,
	}
)

func init() {
	rootCmd.AddCommand(serveCmd)
}

func serveCmdRun(c *cobra.Command, args []string) {
	logLevel := os.Getenv("LOG_LEVEL")
	var level slog.Level
	switch strings.ToLower(logLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr,
		&slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	opts := []web.Option{
		web.WithTrimPrefix(os.Getenv("POTD_TRIM_PREFIX")),
		web.WithTrimSuffix(os.Getenv("POTD_TRIM_SUFFIX")),
	}
	bind := os.Getenv("POTD_ADDR")
	if bind == "" {
		opts = append(opts, web.WithBind(bind))
	}

	if path, ok := os.LookupEnv("POTD_DEBUG"); ok {
		opts = append(opts, web.WithTemplatePath(path))
	}

	caCert := os.Getenv("POTD_CLIENT_CA")
	if caCert == "" {
		caCert = "client-ca.pem"
	}
	opts = append(opts, web.WithClientCA(caCert))

	var d *gorm.DB
	driver := strings.ToLower(os.Getenv("POTD_DB"))
	switch driver {
	case "sqlite":
		fallthrough
	default:
		dbPath := os.Getenv("POTD_SQLITE_PATH")
		if dbPath == "" {
			dbPath = "potd.db"
		}
		var err error
		d, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if err != nil {
			slog.Error("Error opening database", "error", err)
			os.Exit(1)
		}

	}
	d.AutoMigrate(&types.EscrowedToken{})
	opts = append(opts, web.WithDB(d))

	w, err := web.New(opts...)
	if err != nil {
		slog.Error("Error initializing server", "error", err)
		os.Exit(1)
	}
	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				slog.Error("Graceful shutdown timed out.. forcing exit.")
				os.Exit(5)
			}
		}()

		err := w.Shutdown(shutdownCtx)
		if err != nil {
			slog.Error("Error occured during shutdown", "error", err)
		}
		serverStopCtx()

	}()

	cert := os.Getenv("POTD_TLS_CERT")
	if cert == "" {
		cert = "tls.pem"
	}
	key := os.Getenv("POTD_TLS_KEY")
	if key == "" {
		key = "tls.key"
	}
	w.Serve(cert, key)
	<-serverCtx.Done()
}
