package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rakateja/repogen/out/config"
)

type Block func(tx *sqlx.Tx) error

type PostgreSQL struct {
	db *sqlx.DB
}

func NewPostgreSQL(ctx context.Context, conf config.Config) (*PostgreSQL, error) {
	publicConnString := fmt.Sprintf("postgres://%s@%s:%d/%s?sslmode=disable",
		conf.User,
		conf.Host,
		conf.Port,
		conf.Database)

	log.Printf("Connecting to PostgreSQL: %s", publicConnString)
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		conf.User,
		conf.Password,
		conf.Host,
		conf.Port,
		conf.Database)
	db, err := sqlx.Open("postgres", connString)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(conf.MaxOpenConns)
	db.SetMaxIdleConns(conf.MaxIdleConns)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	log.Printf("Connected to PostgreSQL")
	return &PostgreSQL{db: db}, nil
}

func (m *PostgreSQL) WithTransaction(ctx context.Context, block Block) error {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "can't start DB transaction")
	}
	err = block(tx)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return errors.Wrap(err, "rollback fails")
		}
		return err
	}
	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "transaction commit fails")
	}
	return nil
}

func (m *PostgreSQL) In(query string, params map[string]interface{}) (string, []interface{}, error) {
	query, args, err := sqlx.Named(query, params)
	if err != nil {
		return "", nil, err
	}
	return sqlx.In(query, args...)
}

func (m *PostgreSQL) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return m.db.Get(dest, query, args...)
}

func (m *PostgreSQL) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return m.db.Select(dest, query, args...)
}

func (m *PostgreSQL) Rebind(query string) string {
	return m.db.Rebind(query)
}

func (m *PostgreSQL) Query(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	return m.db.Queryx(query, args...)
}

func (m *PostgreSQL) QueryRow(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	return m.db.QueryRowxContext(ctx, query, args...)
}

func (m *PostgreSQL) NamedExec(query string, input map[string]interface{}) (sql.Result, error) {
	return m.db.NamedExec(query, input)
}

func (m *PostgreSQL) PreparedStmt(query string) (*sql.Stmt, error) {
	return m.db.Prepare(query)
}

func (m *PostgreSQL) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return m.db.ExecContext(ctx, query, args...)
}
