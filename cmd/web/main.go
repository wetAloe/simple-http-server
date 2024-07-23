package main

import (
	"crypto/tls"
	"database/sql"
	"errors"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"simple-http-server/internal/models"
	"time"

	_ "github.com/lib/pq"
)

type Config struct {
	addr string
	dsn  string
}

func parseConfig() *Config {
	c := &Config{}
	flag.StringVar(&c.addr, "addr", "8080", "http listen address")
	flag.StringVar(&c.dsn,
		"dsn",
		"postgres://web:123@localhost:5432/snippetbox?sslmode=disable",
		"database connection string",
	)
	flag.Parse()
	return c
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func run(config *Config, logger *slog.Logger) (err error) {
	// Open database
	db, err := openDB(config.dsn)
	if err != nil {
		return err
	}
	defer func(db *sql.DB) {
		closeErr := db.Close()
		if err == nil {
			err = closeErr
		}
	}(db)

	// Open templates
	templates, err := newTemplateCache()
	if err != nil {
		return err
	}

	// Create server
	server := NewServer(
		logger,
		&models.SnippetModel{DB: db},
		&models.UsersModel{DB: db},
		templates,
	)
	httpServer := &http.Server{
		Addr:    net.JoinHostPort("localhost", config.addr),
		Handler: server,
		// to match the code logger
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
		// Initialize a tls.Config struct to hold the non-default TLS settings we
		// want the server to use. In this case the only thing that we're changing
		// is the curve preferences value, so that only elliptic curves with
		// assembly implementations are used.
		TLSConfig: &tls.Config{
			CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
			MinVersion:       tls.VersionTLS10,
			MaxVersion:       tls.VersionTLS12,
		},
		// Add Idle, Read and Write timeouts to the server.
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	defer func(httpServer *http.Server) {
		logger.Info("Shutting down http server")
		closeErr := httpServer.Close()
		if err == nil {
			err = closeErr
		}
	}(httpServer)

	// Start server
	logger.Info("Starting server", "addr", config.addr)
	err = httpServer.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	// err = httpServer.ListenAndServe() for http run
	return err
}

func main() {
	config := parseConfig()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if err := run(config, logger); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
