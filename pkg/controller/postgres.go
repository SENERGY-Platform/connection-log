package controller

import (
	"context"
	"fmt"

	"github.com/SENERGY-Platform/connection-log/pkg/configuration"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultPostgresPort = 5432

// newPostgresPool prepares the pool for the connection history. pgxpool connects lazily, so an unreachable
// database does not fail the start; the first query reports it.
// Credentials are set on the parsed config instead of being formatted into a connection string,
// so special characters in the password need no escaping.
func newPostgresPool(ctx context.Context, config configuration.Config) (*pgxpool.Pool, error) {
	if config.PostgresPort < 0 || config.PostgresPort > 65535 {
		return nil, fmt.Errorf("invalid PostgresPort %d", config.PostgresPort)
	}
	poolConfig, err := pgxpool.ParseConfig("")
	if err != nil {
		return nil, err
	}
	if config.PostgresHost != "" {
		poolConfig.ConnConfig.Host = config.PostgresHost
	}
	poolConfig.ConnConfig.Port = defaultPostgresPort
	if config.PostgresPort != 0 {
		poolConfig.ConnConfig.Port = uint16(config.PostgresPort)
	}
	if config.PostgresUser != "" {
		poolConfig.ConnConfig.User = config.PostgresUser
	}
	poolConfig.ConnConfig.Password = config.PostgresPw
	if config.PostgresDb != "" {
		poolConfig.ConnConfig.Database = config.PostgresDb
	}
	return pgxpool.NewWithConfig(ctx, poolConfig)
}
