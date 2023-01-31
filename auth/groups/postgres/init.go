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
					`CREATE TABLE IF NOT EXISTS groups (
						id          VARCHAR(254) PRIMARY KEY,
						parent_id   VARCHAR(254),
						owner_id    VARCHAR(254) NOT NULL,
						name        VARCHAR(254) NOT NULL,
						description VARCHAR(1024),
						metadata    JSONB,
						created_at  TIMESTAMP,
						updated_at  TIMESTAMP,
						status      SMALLINT NOT NULL CHECK (status >= 0) DEFAULT 1,
						UNIQUE (owner_id, name),
						FOREIGN KEY (parent_id) REFERENCES groups (id) ON DELETE CASCADE
					)`,
				},
				Down: []string{
					`DROP TABLE IF EXISTS groups`,
				},
			},
		},
	}

	_, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	return err
}
