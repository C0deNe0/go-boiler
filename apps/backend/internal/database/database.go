package database

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/C0deNe0/go-boiler/internal/config"
	loggerConfig "github.com/C0deNe0/go-boiler/internal/logger"
	pgxzero "github.com/jackc/pgx-zerolog"
	"github.com/newrelic/go-agent/v3/integrations/nrpgx5"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/rs/zerolog"
)

type Database struct {
	Pool *pgxpool.Pool
	log  *zerolog.Logger
}

type multiTracer struct {
	traces []any
}

const DatabasePingTimeout = 10

//TraceQueryStart implements pgx tracer interface

func (mt *multiTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	for _, tracer := range mt.traces {
		if t, ok := tracer.(interface {
			TracerQueryStart(context.Context, *pgx.Conn, pgx.TraceQueryStartData) context.Context
		}); ok {
			ctx = t.TracerQueryStart(ctx, conn, data)
		}
	}
	return ctx

}

// TraceQueryEnd implements pgx tracer interface

func (mt *multiTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	for _, tracer := range mt.traces {
		if t, ok := tracer.(interface {
			TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData)
		}); ok {
			t.TraceQueryEnd(ctx, conn, data)
		}
	}

}

func New(cfg *config.Config, logger *zerolog.Logger, loggerService *loggerConfig.LoggerService) (*Database, error) {
	hostPort := net.JoinHostPort(cfg.Database.Host, strconv.Itoa(cfg.Database.Port))

	//url encoded password
	encodedPassword := url.QueryEscape(cfg.Database.Password)
	dns := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		cfg.Database.User,
		encodedPassword,
		hostPort,
		cfg.Database.Name,
		cfg.Database.SSLMode, //

	)
	pgxPoolConfig, err := pgxpool.ParseConfig(dns)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx pool config : %w", err)
	}

	//Add if has newrelic postgres instrumental
	if loggerService != nil && loggerService.GetApplication() != nil {
		pgxPoolConfig.ConnConfig.Tracer = nrpgx5.NewTracer()
	}

	if cfg.Primary.Env == "local" {
		globalLever := logger.GetLevel()
		pgxLogger := loggerConfig.NewPgxLogger(globalLever)

		if pgxPoolConfig.ConnConfig.Tracer != nil {
			//if new relic tracer exist, create a multi-tracer

			localTracer := &tracelog.TraceLog{
				Logger:   pgxzero.NewLogger(pgxLogger),
				LogLevel: tracelog.LogLevel(loggerConfig.GetPgxTraceLogLevel(globalLever)),
			}
			pgxPoolConfig.ConnConfig.Tracer = &multiTracer{
				traces: []any{pgxPoolConfig.ConnConfig.Tracer, localTracer},
			}
		} else {
			pgxPoolConfig.ConnConfig.Tracer = &tracelog.TraceLog{
				Logger:   pgxzero.NewLogger(pgxLogger),
				LogLevel: tracelog.LogLevel(loggerConfig.GetPgxTraceLogLevel(globalLever)),
			}
		}

	}

	pool, err := pgxpool.NewWithConfig(context.Background(), pgxPoolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool:%w", err)
	}

	database := &Database{
		Pool: pool,
		log:  logger,
	}

	ctx, cancel := context.WithTimeout(context.Background(), DatabasePingTimeout*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info().Msg("Connected to database ")
	return database, nil
}

func (db *Database) Close() error {
	db.log.Info().Msg("closing database connection")
	db.Pool.Close()
	return nil
}
