package clickhouse_go_2

import (
	"database/sql"
	"github.com/ClickHouse/clickhouse-go/v2"
)

func OpenDB(opt *clickhouse.Options) *sql.DB {
	return clickhouse.OpenDB(opt)
}
