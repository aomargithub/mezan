package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	mezanHttp "github.com/aomargithub/mezan/internal/http"
)



func main() {
	addr := flag.String("addr", ":4000", "Http port")
	staticDir := flag.String("static-dir", "./ui/static", "Path to static assets")
	dsn := flag.String("dsn", "host=localhost port=5454 user=mezan password=mezan dbname=mezan sslmode=disable", "Postgresql data source name")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout , &slog.HandlerOptions{
		Level: slog.LevelInfo,
		AddSource: true,
	}))

	db, err := openDB(*dsn)
    if err != nil {
        logger.Error(err.Error())
        os.Exit(1)
    }
    defer db.Close()

	

	server := mezanHttp.Server{
		Logger: logger,
		Addr: *addr,
		StaticDir: *staticDir,
		Db: db,
	}

	server.Init()
	logger.Info("initializing the serve on", slog.Any("addr", *addr))
	err = http.ListenAndServe(*addr, server.Mux)
	logger.Error(err.Error())
	os.Exit(1)
}

func openDB(dsn string) (*sql.DB, error) {
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        return nil, err
    }

    err = db.Ping()
    if err != nil {
        db.Close()
        return nil, err
    }

    return db, nil
}