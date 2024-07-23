package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"simple-http-server/internal/models"
)

func addRoutes(
	mux *http.ServeMux,
	logger *slog.Logger,
	snippets *models.SnippetModel,
	users *models.UsersModel,
	templates map[string]*template.Template,
) {
	fileServer := http.FileServer(http.Dir("./ui/static/"))

	mux.Handle("GET /{$}", handleHome(logger, snippets, templates))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))
	mux.Handle("GET /snippet/view/{id}", snippetView(logger, snippets, templates))
	mux.Handle("GET /snippet/create", snippetCreateForm(logger, templates))
	mux.Handle("GET /user/signup", userSignup(logger, templates))
	mux.Handle("GET /user/login", userLogin(logger, templates))
	mux.Handle("GET /user/logout", userLogout(logger, users, templates))

	mux.Handle("POST /snippet/create", snippetCreatePost(logger, snippets, templates))
	mux.Handle("POST /user/signup", userSignupPost(logger, users, templates))
	mux.Handle("POST /user/login", userLoginPost(logger, users, templates))
}
