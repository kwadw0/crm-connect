package main

import (
	"context"
	"kwadw0/WhatsCRM/internal/postgres"
	"log/slog"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// @title WhatsCRM API
// @version 1.0
// @description The WhatsCRM API
// @host localhost:3000
// @BasePath /
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	err := godotenv.Load()
	if err != nil {
		logger.Error("Failed to load .env file", "error", err)
		os.Exit(1)
	}

	ttl, err := time.ParseDuration(os.Getenv("TOKEN_TTL"))
	if err != nil {
		logger.Warn("Failed to parse TOKEN_TTL, defaulting to 24h", "error", err)
		ttl = time.Hour * 24
	}

	cfg := config{
		Addr: ":3000",
		db: dbConfig{
			DSN: os.Getenv("DATABASE_URL"),
		},
		jwtSecret: os.Getenv("JWT_SECRET"),
		tokenTTL:  ttl,
	}

	app := application{
		config:    cfg,
		validator: validator.New(),
	}

	logger.Info("Running database migrations...")
    if err := postgres.RunMigrations(cfg.db.DSN); err != nil {
        logger.Error("Migration failed", "error", err)
        os.Exit(1)
    }
    logger.Info("Migrations completed successfully")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := pgxpool.New(ctx, cfg.db.DSN)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	logger.Info("Connected to database", "conn", conn)
	defer conn.Close()

	app.db = conn

    if err := app.run(app.mount()); err != nil {
		logger.Error("Failed to run application", "error", err)
		os.Exit(1)
	}
}
