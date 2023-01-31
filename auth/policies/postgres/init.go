package postgres

import (
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // required for SQL access
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
				Id: "auth_01",
				Up: []string{
					`CREATE TABLE IF NOT EXISTS policies (
						owner_id    VARCHAR(254) NOT NULL,
						subject     VARCHAR(254) NOT NULL,
						object      VARCHAR(254) NOT NULL,
						actions     TEXT[] NOT NULL,
						created_at  TIMESTAMP,
						updated_at  TIMESTAMP,
						FOREIGN KEY (subject) REFERENCES users (id),
						PRIMARY KEY (subject, object, actions)
					)`,
				},
				Down: []string{
					`DROP TABLE IF EXISTS policies`,
				},
			},
		},
	}

	_, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	return err
}
