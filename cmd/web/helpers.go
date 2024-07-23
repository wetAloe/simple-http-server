package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
)

func serverError(w http.ResponseWriter, logger *slog.Logger, err error, args ...any) {
	logger.Error(err.Error(), args...)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func getTemplate(templates map[string]*template.Template, page string) *template.Template {
	ts, ok := templates[page]
	if !ok {
		keys := make([]string, 0, len(templates))
		for key, _ := range templates {
			keys = append(keys, key)
		}
		panic(errors.New(fmt.Sprintf("no template found for %s, available %v", page, keys)))
	}
	return ts
}

func render(w http.ResponseWriter, ts *template.Template, status int, data templateData) error {
	buf := new(bytes.Buffer)
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		return err
	}

	w.WriteHeader(status)
	_, err = buf.WriteTo(w)
	return err
}
