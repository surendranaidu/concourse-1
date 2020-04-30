package db

import (
	"database/sql"
	"strings"

	"code.cloudfoundry.org/lager"
	"github.com/Masterminds/squirrel"
)

func Log(logger lager.Logger, conn Conn) Conn {
	return &logConn{
		Conn:   conn,
		logger: logger,
	}
}

type logConn struct {
	Conn

	logger lager.Logger
}

type logDbTx struct {
	Tx

	logger lager.Logger
}

func (t *logDbTx) QueryRow(query string, args ...interface{}) squirrel.RowScanner {
	t.logger.Debug("tx-query-row", lager.Data{"query": strip(query)})
	return t.Tx.QueryRow(query, args...)
}

func (c *logConn) Begin() (Tx, error) {
	tx, err := c.Conn.Begin()
	if err != nil {
		return nil, err
	}

	return &logDbTx{Tx: tx, logger: c.logger}, nil
}

func (c *logConn) Query(query string, args ...interface{}) (*sql.Rows, error) {
	c.logger.Debug("query", lager.Data{"query": strip(query)})
	return c.Conn.Query(query, args...)
}

func (c *logConn) QueryRow(query string, args ...interface{}) squirrel.RowScanner {
	c.logger.Debug("query-row", lager.Data{"query": strip(query)})
	return c.Conn.QueryRow(query, args...)
}

func (c *logConn) Exec(query string, args ...interface{}) (sql.Result, error) {
	c.logger.Debug("exec", lager.Data{"query": strip(query)})
	return c.Conn.Exec(query, args...)
}

func strip(query string) string {
	return strings.Join(strings.Fields(query), " ")
}
