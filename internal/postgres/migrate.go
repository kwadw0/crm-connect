package postgres

import (
	"database/sql"
	"embed"

	_ "github.com/jackc/pgx/v5/stdlib" // Import the pgx driver for migrations
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func RunMigrations(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	// 2. Point Goose to our embedded filesystem
	goose.SetBaseFS(embedMigrations)

	// 3. Set the dialect to postgres
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	// 4. Run 'Up' migrations
	return goose.Up(db, "migrations")
}
