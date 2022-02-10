package feed

import (
	"database/sql"
	"embed"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
)

func GetDB(config *DBConfig) *sql.DB {
	var pgdb *sql.DB

	dbConfig, err := pgx.ParseConfig(config.ConnStr())
	if err != nil {
		panic("failed to generate database config")
	}

	pgdb = stdlib.OpenDB(*dbConfig)

	if err = pgdb.Ping(); err != nil {
		panic(err)
	}

	pgdb.SetConnMaxLifetime(config.ConnLifetime)
	pgdb.SetMaxOpenConns(config.MaxConns)
	pgdb.SetMaxIdleConns(config.MaxConns)

	if err := MigrateDB(pgdb); err != nil {
		panic(err)
	}

	return pgdb
}

type DBConfig struct {
	User         string
	Password     string
	Host         string
	Port         int
	Database     string
	SSLMode      string
	ConnLifetime time.Duration
	MaxConns     int
}

func (c *DBConfig) ConnStr() string {
	res := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode)

	return res
}

func (c *DBConfig) URLConnStr() string {
	res := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, url.QueryEscape(c.Password), c.Host, c.Port, c.Database, c.SSLMode)

	return res
}

//go:embed migrations
var migrations embed.FS

// version defines the current migration version. This ensures the app
// is always compatible with the version of the database.
const migrationVersion = 10

// Migrate migrates the Postgres schema to the current version.
func MigrateDB(db *sql.DB) error {
	sourceInstance, err := httpfs.New(http.FS(migrations), "migrations")
	if err != nil {
		return err
	}
	targetInstance, err := postgres.WithInstance(db, new(postgres.Config))
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("httpfs", sourceInstance, "postgres", targetInstance)
	if err != nil {
		return err
	}
	err = m.Migrate(migrationVersion) // current version
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return sourceInstance.Close()
}
