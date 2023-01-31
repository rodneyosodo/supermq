package postgres

import (
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib" // required for SQL access
	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux/internal/postgres"
	migrate "github.com/rubenv/sql-migrate"
)

// Connect creates a connection to the PostgreSQL instance and applies any
// unapplied database migrations. A non-nil error is returned to indicate
// failure.
func Connect(cfg postgres.Config) (*sqlx.DB, error) {
	url := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s sslcert=%s sslkey=%s sslrootcert=%s", cfg.Host, cfg.Port, cfg.User, cfg.Name, cfg.Pass, cfg.SSLMode, cfg.SSLCert, cfg.SSLKey, cfg.SSLRootCert)
	db, err := sqlx.Open("pgx", url)
	if err != nil {
		return nil, err
	}

	if err := migrateDB(db); err != nil {
		return nil, err
	}
	return db, nil
}

func migrateDB(db *sqlx.DB) error {
	migrations := &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{
				Id: "clients_01",
				Up: []string{
					`CREATE TABLE IF NOT EXISTS users (
						id          VARCHAR(254) PRIMARY KEY,
						name        VARCHAR(254),
						owner       VARCHAR(254),
						identity    VARCHAR(254) UNIQUE NOT NULL,
						secret      TEXT NOT NULL,
						tags        TEXT[],
						metadata    JSONB,
						created_at  TIMESTAMP,
						updated_at  TIMESTAMP,
						status      SMALLINT NOT NULL CHECK (status >= 0) DEFAULT 1
					)`,
				},
				Down: []string{
					`DROP TABLE IF EXISTS users`,
				},
			},
		},
	}

	_, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	return err
}
