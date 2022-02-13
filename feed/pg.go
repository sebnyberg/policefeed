package feed

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
)

const (
	connectionBaseTimeoutSeconds = 10
)

// DBConfig contains database-releated config settings
// Note: naming in this config shares namespace with the global config,
// hence the "DB" and "PG" prefixes for its keys.
type DBConfig struct {
	DBConnMaxLifetime string `usage:"max lifetime for a db connection, duration format e.g. '10s', '1m'" value:"300s"`
	DBMaxIdleConns    int    `usage:"max idle database connections" value:"10"`
	DBMaxOpenConns    int    `usage:"max open database connections" value:"10"`
	PGDatabase        string `name:"pgdb" env:"PGDATABASE"`
	PGHost            string `name:"pghost" env:"PGHOST"`
	PGPassword        string `name:"pgpassword" env:"PGPASSWORD"`
	PGPort            int    `name:"pgport" env:"PGPORT" value:"5432"`
	PGSSLMode         string `name:"pgsslmode" env:"PGSSLMODE"`
	PGUser            string `name:"pguser" env:"PGUSER"`
}

// URL fetches the pgconfig as a connection string url
func (c *DBConfig) URL() *url.URL {
	var user *url.Userinfo
	if c.PGUser != "" {
		if c.PGPassword != "" {
			user = url.UserPassword(c.PGUser, c.PGPassword)
		} else {
			user = url.User(c.PGUser)
		}
	}
	u := &url.URL{
		Host:   c.PGHost + ":" + strconv.Itoa(c.PGPort),
		Scheme: "postgres",
		User:   user,
		Path:   c.PGDatabase,
	}
	params := url.Values{}
	if c.PGSSLMode != "" {
		params.Add("sslmode", c.PGSSLMode)
	}
	u.RawQuery = params.Encode()
	return u
}

// OpenDB opens a new sql.DB using pgx's stdlib bindings
func (c *DBConfig) OpenDB() (*sql.DB, error) {
	cfg, err := pgx.ParseConfig(c.URL().String())
	if err != nil {
		return nil, err
	}
	db := stdlib.OpenDB(*cfg)
	db.SetMaxOpenConns(c.DBMaxOpenConns)
	db.SetMaxIdleConns(c.DBMaxIdleConns)
	connLifetime, err := time.ParseDuration(c.DBConnMaxLifetime)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to parse database connection lifetime: %s", err.Error())
	}
	db.SetConnMaxLifetime(connLifetime)
	pingCtx, cancelPingCtx := context.WithTimeout(
		context.Background(), connectionBaseTimeoutSeconds*time.Second)
	defer cancelPingCtx()
	if err := db.PingContext(pingCtx); err != nil {
		return nil, errors.New("failed to ping database")
	}
	return db, nil
}

//go:embed migrations
var migrations embed.FS

// version defines the current migration version. This ensures the app
// is always compatible with the version of the database.
const migrationVersion = 1

// Migrate migrates the Postgres schema to the current version.
func ValidateSchema(db *sql.DB) error {
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
