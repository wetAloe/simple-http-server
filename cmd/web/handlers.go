package main

import (
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"simple-http-server/internal/models"
	"simple-http-server/internal/validator"
	"strconv"
)

type createForm struct {
	Title    string
	Content  string
	Expires  int
	Problems map[string]string
}

func (f *createForm) Validate() map[string]string {
	f.Problems = make(map[string]string)

	if !validator.NotBlank(f.Title) {
		f.Problems["title"] = "Title field cannot be blank"
	} else if !validator.MaxRunes(f.Title, 100) {
		f.Problems["title"] = "Title field too long"
	}

	if !validator.NotBlank(f.Content) {
		f.Problems["content"] = "Content field cannot be blank"
	}

	if !validator.PermittedValue(f.Expires, 1, 7, 365) {
		f.Problems["expires"] = "Expires field must equal 1 or 7 or 365"
	}

	return f.Problems
}

type userSignupForm struct {
	Name     string            `form:"name"`
	Email    string            `form:"email"`
	Password string            `form:"password"`
	Problems map[string]string `form:"-"`
}

func (f *userSignupForm) Validate() map[string]string {
	f.Problems = make(map[string]string)

	if !validator.NotBlank(f.Name) {
		f.Problems["name"] = "Name field cannot be blank"
	}

	if !validator.NotBlank(f.Email) {
		f.Problems["email"] = "Email field cannot be blank"
	} else if !validator.Matches(f.Email, validator.EmailRX) {
		f.Problems["email"] = "Provided email has invalid format"
	}

	if !validator.NotBlank(f.Password) {
		f.Problems["password"] = "Password field cannot be blank"
	} else if !validator.MinChars(f.Password, 8) {
		f.Problems["password"] = "Password has to be longer than 8 characters"
	}

	return f.Problems
}

type userLoginForm struct {
	Email    string            `form:"email"`
	Password string            `form:"password"`
	Problems map[string]string `form:"-"`
}

// *GET*

func handleHome(
	logger *slog.Logger,
	snippets *models.SnippetModel,
	templates map[string]*template.Template,
) http.Handler {
	ts := getTemplate(templates, "home.tmpl.html")

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			latestSnippets, err := snippets.Latest()
			if err != nil {
				serverError(w, logger, err, "method", r.Method, "uri", r.URL.RequestURI())
			}

			data := newTemplateData()
			data.Snippets = latestSnippets
			if err := render(w, ts, http.StatusOK, data); err != nil {
				serverError(w, logger, err, "method", r.Method, "uri", r.URL.RequestURI())
				return
			}
		})
}

func snippetView(
	logger *slog.Logger,
	snippets *models.SnippetModel,
	templates map[string]*template.Template,
) http.Handler {
	ts := getTemplate(templates, "view.tmpl.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil || id < 1 {
			http.NotFound(w, r)
			return
		}

		snippet, err := snippets.Get(id)
		if err != nil {
			if errors.Is(err, models.ErrNoRecord) {
				http.NotFound(w, r)
			} else {
				serverError(w, logger, err, "method", r.Method, "uri", r.URL.RequestURI())
			}
			return
		}

		data := newTemplateData()
		data.Snippet = snippet
		if err := render(w, ts, http.StatusOK, data); err != nil {
			serverError(w, logger, err, "method", r.Method, "uri", r.URL.RequestURI())
			return
		}
	})
}

func snippetCreateForm(
	logger *slog.Logger,
	templates map[string]*template.Template,
) http.Handler {
	ts := getTemplate(templates, "create.tmpl.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := newTemplateData()
		data.Form = createForm{Expires: 365}

		if err := render(w, ts, http.StatusOK, data); err != nil {
			serverError(w, logger, err, "method", r.Method, "uri", r.URL.RequestURI())
			return
		}
	})
}

func userSignup(
	logger *slog.Logger,
	templates map[string]*template.Template,
) http.Handler {
	ts := getTemplate(templates, "signup.tmpl.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := newTemplateData()
		data.Form = userSignupForm{}
		if err := render(w, ts, http.StatusOK, data); err != nil {
			serverError(w, logger, err, "method", r.Method, "uri", r.URL.RequestURI())
			return
		}
	})
}

func userLogin(
	logger *slog.Logger,
	templates map[string]*template.Template,
) http.Handler {
	ts := getTemplate(templates, "login.tmpl.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := newTemplateData()
		data.Form = userLoginForm{}
		if err := render(w, ts, http.StatusOK, data); err != nil {
			serverError(w, logger, err, "method", r.Method, "uri", r.URL.RequestURI())
			return
		}
	})
}

func userLogout(
	logger *slog.Logger,
	users *models.UsersModel,
	templates map[string]*template.Template,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}

// *POST*

func snippetCreatePost(
	logger *slog.Logger,
	snippets *models.SnippetModel,
	templates map[string]*template.Template,
) http.Handler {
	ts := getTemplate(templates, "create.tmpl.html")

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				clientError(w, http.StatusBadRequest)
				return
			}

			expires, err := strconv.Atoi(r.PostForm.Get("expires"))
			if err != nil {
				clientError(w, http.StatusBadRequest)
				return
			}
			form := createForm{
				Title:   r.PostForm.Get("title"),
				Content: r.PostForm.Get("content"),
				Expires: expires,
			}

			problems := form.Validate()
			if len(problems) > 0 {
				data := newTemplateData()
				data.Form = form

				err := render(w, ts, http.StatusUnprocessableEntity, data)
				if err != nil {
					serverError(w, logger, err, "method", r.Method, "uri", r.URL.RequestURI())
				}
				return
			}

			id, err := snippets.Insert(form.Title, form.Content, form.Expires)
			if err != nil {
				serverError(w, logger, err, "method", r.Method, "uri", r.URL.RequestURI())
				return
			}

			http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
		})
}

func userSignupPost(
	logger *slog.Logger,
	users *models.UsersModel,
	templates map[string]*template.Template,
) http.Handler {
	ts := getTemplate(templates, "signup.tmpl.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			clientError(w, http.StatusBadRequest)
			return
		}

		form := userSignupForm{
			Name:     r.PostForm.Get("name"),
			Email:    r.PostForm.Get("email"),
			Password: r.PostForm.Get("password"),
		}
		problems := form.Validate()
		if len(problems) > 0 {
			data := newTemplateData()
			data.Form = form

			err := render(w, ts, http.StatusUnprocessableEntity, data)
			if err != nil {
				serverError(w, logger, err, "method", r.Method, "uri", r.URL.RequestURI())
			}
			return
		}

		err = users.Insert(form.Name, form.Email, form.Password)
		if err != nil {
			if errors.Is(err, models.ErrDuplicateEmail) {
				form.Problems["email"] = "Provided email already in use"

				data := newTemplateData()
				data.Form = form
				err := render(w, ts, http.StatusUnprocessableEntity, data)
				if err != nil {
					serverError(w, logger, err, "method", r.Method, "uri", r.URL.RequestURI())
				}
			} else {
				serverError(w, logger, err, "method", r.Method, "uri", r.URL.RequestURI())
			}
			return
		}

		// And redirect the user to the login page.
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
	})
}

func userLoginPost(
	logger *slog.Logger,
	users *models.UsersModel,
	templates map[string]*template.Template,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}
