package postgres

import (
	"database/sql"
	"embed"
	"errors"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(databaseURL string) error {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to open postgres connection for migration", err)
	}
	defer db.Close()

	driver, err := pgxmigrate.WithInstance(db, &pgxmigrate.Config{})
	if err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to initialize migration driver", err)
	}

	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to load embedded migrations", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "pgx5", driver)
	if err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to initialize migrator", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return apperror.Wrap(apperror.CodeInternal, "failed to run migrations", err)
	}
	return nil
}