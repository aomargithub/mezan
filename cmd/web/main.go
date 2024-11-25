package main

import (
	"crypto/tls"
	"flag"
	mezanHttp "github.com/aomargithub/mezan/internal/http"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	addr := flag.String("addr", ":4000", "Http port")
	staticDir := flag.String("static-dir", "./ui/static", "Path to static assets")
	dsn := flag.String("dsn", "host=localhost port=5454 user=mezan password=mezan dbname=mezan sslmode=disable", "Postgresql data source name")
	flag.Parse()

	server := mezanHttp.Server{
		StaticDir: *staticDir,
		DSN:       *dsn,
	}

	server.Init()
	defer server.DB.Close()
	server.Logger.Info("initializing the serve on", slog.Any("addr", *addr))

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:         *addr,
		Handler:      server.Mux,
		ErrorLog:     slog.NewLogLogger(server.Logger.Handler(), slog.LevelError),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	err := srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	server.Logger.Error(err.Error())
	os.Exit(1)
}
