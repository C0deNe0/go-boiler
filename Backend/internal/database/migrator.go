package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/url"
	"strconv"

	"github.com/C0deNe0/go-boiler/internal/config"
	"github.com/jackc/pgx/v5"
	tern "github.com/jackc/tern/v2/migrate"
	"github.com/rs/zerolog"
)

// go:Embed migrations/*.sql
var migrations embed.FS

func Migrate(ctx context.Context, looger *zerolog.Logger, cfg *config.Config) error {
	hostPort := net.JoinHostPort(cfg.Database.Host, strconv.Itoa(cfg.Database.Port))

	encodedPassword := url.QueryEscape(cfg.Database.Password)
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		cfg.Database.Host,
		encodedPassword,
		hostPort,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return err
	}

	defer conn.Close(ctx)
	m, err := tern.NewMigrator(ctx, conn, "schema_version")
	if err != nil {
		return fmt.Errorf("constructing database migrator:%w", err)
	}

	subtree, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("retreving database migration subtree:%w", err)
	}

	if err := m.LoadMigrations(subtree); err != nil {
		return fmt.Errorf("Loading database migrations: %w", err)
	}

	from, err := m.GetCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("retriving current database migration version")
	}

	if err := m.Migrate(ctx); err != nil {
		return err
	}

	if from == int32(len(m.Migrations)) {
		looger.Info().Msgf("database schema up to date, version %d", len(m.Migrations))
	} else {
		looger.Info().Msgf("migrated database schema, from %d to %d", from, len(m.Migrations))
	}
	return nil
}
