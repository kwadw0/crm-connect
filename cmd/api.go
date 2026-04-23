package main

import (
	"kwadw0/WhatsCRM/auth"
	"kwadw0/WhatsCRM/internal/postgres/repo"
	"kwadw0/WhatsCRM/organizations"
	"kwadw0/WhatsCRM/roles"
	"kwadw0/WhatsCRM/users"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "kwadw0/WhatsCRM/docs"
)


func (app *application) run (h http.Handler) error {
	slog.Info("Server started on ", app.config.Addr, h)
	return http.ListenAndServe(app.config.Addr, h)
}

func (app *application) mount () http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	// Serve generated Swaggo Docs at /docs
	r.Get("/docs/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:3000/docs/doc.json"), //The url pointing to API definition
	))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})
	userService := users.NewService(repo.New(app.db))
	userHandler := users.NewHandler(userService)

	// Grouping all /users endpoints together
	r.Route("/users", func(r chi.Router) {
		r.Post("/", userHandler.CreateUser)
		r.Get("/", userHandler.GetAllUsers)
		r.Get("/{id}", userHandler.GetUserByID)
		r.Put("/{id}", userHandler.UpdateUser)
		r.Delete("/{id}", userHandler.DeleteUser)
	})

	authService := auth.NewService(repo.New(app.db),  []byte(app.config.jwtSecret), app.config.tokenTTL)
	authHandler := auth.AuthHandler(authService, app.validator)
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.RegisterUser)
		r.Post("/login", authHandler.LoginUser)
	})

	roleService := roles.NewService(repo.New(app.db))
	roleHandler := roles.NewHandler(roleService, app.validator)
	r.Route("/roles", func(r chi.Router) {
		r.Post("/", roleHandler.CreateRole)
		r.Get("/", roleHandler.GetAllRoles)
		r.Get("/{id}", roleHandler.GetRoleByID)
		r.Put("/{id}", roleHandler.UpdateRole)
		r.Delete("/{id}", roleHandler.DeleteRole)
	})

	// --- PROTECTED ROUTES (Requires AuthMiddleware) ---
	r.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddleware([]byte(app.config.jwtSecret)))

		orgService := organizations.NewOrganizationService(repo.New(app.db))
		orgHandler := organizations.NewOrganizationHandler(orgService, userService, app.validator)

		// Grouping all /organizations endpoints together
		r.Route("/organizations", func(r chi.Router) {
			r.Post("/", orgHandler.AddOrganization)
			r.Get("/", orgHandler.ListOrganizations)
			r.Get("/{id}", orgHandler.GetOrganizationByID)
			r.Get("/current", orgHandler.GetCurrentUserOrganization)
			r.Put("/{id}", orgHandler.UpdateOrganization)
			r.Delete("/{id}", orgHandler.DeleteOrganization)
		})
	})

	return r	
}


type application struct {
	config    config
	db        *pgxpool.Pool
	validator *validator.Validate
}


type config struct {
	Addr string
	db dbConfig
	jwtSecret string
	tokenTTL  time.Duration
}

type dbConfig struct {
	DSN string
}

