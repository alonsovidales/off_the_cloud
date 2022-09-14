package models

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var pool *sql.DB

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var err error
	pass := os.Getenv("MYSQL_PASS")
	pool, err = sql.Open("mysql", fmt.Sprintf("root:%s@/otc_test?parseTime=true&loc=UTC", pass))

	if err != nil {
		panic(fmt.Sprintf("Error connecting to MySQL: %s", err))
	}

	pool.SetConnMaxLifetime(time.Minute * 3)
	pool.SetMaxOpenConns(10)
	pool.SetMaxIdleConns(10)
}

func execStm(statement string, args ...any) (sql.Result, error) {
	stmt, err := pool.Prepare(statement)
	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()
	return stmt.Exec(args...)
}

func getRow(ctx context.Context, query string, args ...any) *sql.Row {
	return pool.QueryRowContext(ctx, query, args...)
}

func ClosePool() error {
	return pool.Close()
}
