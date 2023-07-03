// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux/internal/clients/postgres"
	"github.com/mainflux/mainflux/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ErrFailedToLookupIP is returned when the IP address of the database peer cannot be looked up.
var ErrFailedToLookupIP = errors.New("failed to lookup IP address")

var _ Database = (*database)(nil)

type database struct {
	db     *sqlx.DB
	tracer trace.Tracer
	dbUser string
	dbName string
	dbHost string
	dbPort string
	dbIPV4 string
	dbIPV6 string
}

// Database provides a database interface.
type Database interface {
	// NamedQueryContext executes a named query against the database and returns
	NamedQueryContext(context.Context, string, interface{}) (*sqlx.Rows, error)

	// NamedExecContext executes a named query against the database and returns
	NamedExecContext(context.Context, string, interface{}) (sql.Result, error)

	// QueryRowxContext queries the database and returns an *sqlx.Row
	QueryRowxContext(context.Context, string, ...interface{}) *sqlx.Row

	// QueryxContext queries the database and returns an *sqlx.Rows
	QueryxContext(context.Context, string, ...interface{}) (*sqlx.Rows, error)

	// QueryContext executes a query that returns rows, typically a SELECT.
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)

	// ExecContext executes a query without returning any rows
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)

	// BeginTxx begins a transaction and returns an *sqlx.Tx
	BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
}

// NewDatabase creates a Clients'Database instance.
func NewDatabase(db *sqlx.DB, config postgres.Config, tracer trace.Tracer) (Database, error) {
	database := &database{
		db:     db,
		tracer: tracer,
		dbUser: config.User,
		dbName: config.Name,
		dbHost: config.Host,
		dbPort: config.Port,
	}

	ipAddrs, err := net.LookupIP(config.Host)
	if err != nil {
		return database, errors.Wrap(ErrFailedToLookupIP, err)
	}

	for _, ipv4Addr := range ipAddrs {
		if ipv4Addr.To4() != nil {
			database.dbIPV4 = ipv4Addr.String()
		}
	}
	for _, ipv6Addr := range ipAddrs {
		if ipv6Addr.To16() != nil && ipv6Addr.To4() == nil {
			database.dbIPV6 = ipv6Addr.String()
		}
	}

	return database, nil
}

func (d *database) NamedQueryContext(ctx context.Context, query string, args interface{}) (*sqlx.Rows, error) {
	ctx, span := d.addSpanTags(ctx, "NamedQueryContext", query)
	defer span.End()

	return d.db.NamedQueryContext(ctx, query, args)
}

func (d *database) NamedExecContext(ctx context.Context, query string, args interface{}) (sql.Result, error) {
	ctx, span := d.addSpanTags(ctx, "NamedExecContext", query)
	defer span.End()

	return d.db.NamedExecContext(ctx, query, args)
}

func (d *database) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	ctx, span := d.addSpanTags(ctx, "ExecContext", query)
	defer span.End()

	return d.db.ExecContext(ctx, query, args...)
}

func (d *database) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	ctx, span := d.addSpanTags(ctx, "QueryRowxContext", query)
	defer span.End()

	return d.db.QueryRowxContext(ctx, query, args...)
}

func (d *database) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	ctx, span := d.addSpanTags(ctx, "QueryxContext", query)
	defer span.End()

	return d.db.QueryxContext(ctx, query, args...)
}

func (d database) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	ctx, span := d.addSpanTags(ctx, "QueryContext", query)
	defer span.End()
	return d.db.QueryContext(ctx, query, args...)
}

func (d database) BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	ctx, span := d.addSpanTags(ctx, "BeginTxx", "")
	defer span.End()

	return d.db.BeginTxx(ctx, opts)
}

func (d *database) addSpanTags(ctx context.Context, method, query string) (context.Context, trace.Span) {
	var address string
	if d.dbHost != "" && d.dbPort != "" {
		address = d.dbHost + ":" + d.dbPort
	}
	if d.dbHost != "" && d.dbPort == "" {
		address = d.dbHost
	}

	ctx, span := d.tracer.Start(ctx,
		fmt.Sprintf("sql_%s", method),
		trace.WithAttributes(
			attribute.String("db.type", "sql"),
			attribute.String("db.instance", d.dbName),
			attribute.String("db.user", d.dbUser),
			attribute.String("db.statement", query),
			attribute.String("span.kind", "client"),
			attribute.String("peer.address", address),
			attribute.String("peer.hostname", d.dbHost),
			attribute.String("peer.ipv4", d.dbIPV4),
			attribute.String("peer.ipv6", d.dbIPV6),
			attribute.String("peer.port", d.dbPort),
			attribute.String("peer.service", "postgres"),
		),
	)

	return ctx, span
}
