package connect

import (
	"context"
	databasesql "database/sql"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

type CHConnection struct {
	params *CHConnectionParams
	conn   *databasesql.DB
}

func NewConnection(params *CHConnectionParams) *CHConnection {
	// DO not perform connection immediately, do it in lazy manner
	return &CHConnection{
		params: params,
	}
}

func (c *CHConnection) connect() {

	logrus.Infof("Establishing connection: %s", c.params.GetDSNWithHiddenCredentials())
	dbConnection, err := databasesql.Open("clickhouse", c.params.GetDSN())
	if err != nil {
		logrus.Errorf("FAILED Open(%s) %v", c.params.GetDSNWithHiddenCredentials(), err)
		return
	}

	// Ping should be deadlined
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(defaultTimeout))
	defer cancel()

	if err := dbConnection.PingContext(ctx); err != nil {
		logrus.Errorf("FAILED Ping(%s) %v", c.params.GetDSNWithHiddenCredentials(), err)
		_ = dbConnection.Close()
		return
	}

	c.conn = dbConnection
}

func (c *CHConnection) ensureConnected() bool {
	if c.conn != nil {
		logrus.Infof("Already connected: %s", c.params.GetDSNWithHiddenCredentials())
		return true
	}

	c.connect()

	return c.conn != nil
}

// Query
type Query struct {
	ctx        context.Context
	cancelFunc context.CancelFunc

	Rows *databasesql.Rows
}

// Close
func (q *Query) Close() {
	if q == nil {
		return
	}

	if q.Rows != nil {
		err := q.Rows.Close()
		q.Rows = nil
		if err != nil {
			logrus.Infof("UNABLE to close rows. err: %v", err)
		}
	}

	if q.cancelFunc != nil {
		q.cancelFunc()
		q.cancelFunc = nil
	}
}

// Query runs given sql query
func (c *CHConnection) Query(sql string) (*Query, error) {
	if len(sql) == 0 {
		return nil, nil
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(defaultTimeout))

	if !c.ensureConnected() {
		cancel()
		s := fmt.Sprintf("FAILED connect(%s) for SQL: %s", c.params.GetDSNWithHiddenCredentials(), sql)
		logrus.Error(s)
		return nil, fmt.Errorf(s)
	}

	rows, err := c.conn.QueryContext(ctx, sql)
	if err != nil {
		cancel()
		s := fmt.Sprintf("FAILED Query(%s) %v for SQL: %s", c.params.GetDSN(), err, sql)
		logrus.Error(s)
		return nil, err
	}

	logrus.Infof("clickhouse.QueryContext():'%s'", sql)

	return &Query{
		ctx:        ctx,
		cancelFunc: cancel,
		Rows:       rows,
	}, nil
}

// Exec runs given sql query
func (c *CHConnection) Exec(sql string) error {
	if len(sql) == 0 {
		return nil
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(defaultTimeout))
	defer cancel()

	if !c.ensureConnected() {
		s := fmt.Sprintf("FAILED connect(%s) for SQL: %s", c.params.GetDSNWithHiddenCredentials(), sql)
		logrus.Errorf(s)
		return fmt.Errorf(s)
	}

	_, err := c.conn.ExecContext(ctx, sql)

	if err != nil {
		logrus.Errorf("FAILED Exec(%s) %v for SQL: %s", c.params.GetDSNWithHiddenCredentials(), err, sql)
		return err
	}

	logrus.Infof("clickhouse.Exec(%s) for SQL: %s", c.params.GetDSNWithHiddenCredentials(), sql)

	return nil
}
