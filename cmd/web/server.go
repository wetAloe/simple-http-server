package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"simple-http-server/internal/models"
)

func NewServer(
	logger *slog.Logger,
	snippets *models.SnippetModel,
	users *models.UsersModel,
	templates map[string]*template.Template,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, logger, snippets, users, templates)

	var handler http.Handler = mux
	handler = commonHeaders(handler)
	handler = logRequest(logger, handler)
	handler = recoverPanic(logger, handler)
	return handler
}
