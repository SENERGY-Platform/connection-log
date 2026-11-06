Provides HTTP-API to request current and historical connection log. 
Historical data is read from the database `connectionlog` of a TimescaleDB (PostgreSQL) instance, tables `device` and `gateway`
with the columns `id`, `time` and `connected`. The tables are created by the connectionlog-worker service, which also writes them;
this service only reads them.
Current data is read form a mongodb instance.

Configuration of the history database (`config.json` or environment):

| Field          | Environment     | Default         |
|----------------|-----------------|-----------------|
| `PostgresHost` | `POSTGRES_HOST` | `timescale`     |
| `PostgresPort` | `POSTGRES_PORT` | `5432`          |
| `PostgresUser` | `POSTGRES_USER` | `connectionlog` |
| `PostgresPw`   | `POSTGRES_PW`   |                 |
| `PostgresDb`   | `POSTGRES_DB`   | `connectionlog` |

Generate swagger docs (the `go:generate` lines at the end of `pkg/api/handler.go`):

    go generate ./...

Tests of the history queries start TimescaleDB in a container (`timescale/timescaledb-ha:pg18.6-ts2.29.2`) and are skipped when Docker is not available.
