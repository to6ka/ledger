package sqlstorage

import (
	"context"
	"database/sql"
	"github.com/huandu/go-sqlbuilder"
	"github.com/numary/go-libs/sharedapi"
	"math"
	"time"

	"github.com/numary/ledger/pkg/core"
	"github.com/numary/ledger/pkg/ledger/query"
)

func (s *Store) transactionsQuery(p map[string]interface{}) *sqlbuilder.SelectBuilder {

	sb := sqlbuilder.NewSelectBuilder()
	sb.GroupBy("t.id", "t.postings", "t.metadata", "t.timestamp", "t.reference")
	sb.Select("t.id", "t.timestamp", "t.reference", "t.metadata", "t.postings")
	switch s.schema.Flavor() {
	case sqlbuilder.PostgreSQL:
		sb.From(s.schema.Table("transactions")+" t", "jsonb_to_recordset(t.postings) as postings(source varchar, destination varchar, asset varchar, amount bigint)")
	case sqlbuilder.SQLite:
		sb.From("transactions t", "json_each(postings)")
	}
	if account, ok := p["account"]; ok && account.(string) != "" {
		switch s.schema.Flavor() {
		case sqlbuilder.PostgreSQL:
			sb.Where(sb.Or(
				sb.Equal("source", account.(string)),
				sb.Equal("destination", account.(string)),
			))
		case sqlbuilder.SQLite:
			sb.Where(sb.Or(
				sb.Equal("json_extract(json_each.value, '$.source')", account.(string)),
				sb.Equal("json_extract(json_each.value, '$.destination')", account.(string)),
			))
		}
	}
	if ref, ok := p["reference"]; ok && p["reference"].(string) != "" {
		sb.Where(sb.E("reference", ref.(string)))
	}
	return sb
}

func (s *Store) findTransactions(ctx context.Context, exec executor, q query.Query) (sharedapi.Cursor, error) {
	q.Limit = int(math.Max(-1, math.Min(float64(q.Limit), 100))) + 1

	c := sharedapi.Cursor{}

	sb := s.transactionsQuery(q.Params)
	sb.OrderBy("t.id desc")
	if q.After != "" {
		sb.Where(sb.LessThan("t.id", q.After))
	}
	sb.Limit(q.Limit)

	sqlq, args := sb.BuildWithFlavor(s.schema.Flavor())
	rows, err := exec.QueryContext(ctx, sqlq, args...)
	if err != nil {
		return c, s.error(err)
	}
	defer rows.Close()

	transactions := make([]core.Transaction, 0)

	for rows.Next() {
		var (
			ref sql.NullString
			ts  sql.NullString
		)

		tx := core.Transaction{}
		err := rows.Scan(
			&tx.ID,
			&ts,
			&ref,
			&tx.Metadata,
			&tx.Postings,
		)
		if err != nil {
			return c, err
		}
		tx.Reference = ref.String
		if tx.Metadata == nil {
			tx.Metadata = core.Metadata{}
		}
		timestamp, err := time.Parse(time.RFC3339, ts.String)
		if err != nil {
			return sharedapi.Cursor{}, err
		}
		tx.Timestamp = timestamp.UTC().Format(time.RFC3339)
		transactions = append(transactions, tx)
	}
	if rows.Err() != nil {
		return sharedapi.Cursor{}, s.error(err)
	}

	c.PageSize = q.Limit - 1
	c.HasMore = len(transactions) == q.Limit
	if c.HasMore {
		transactions = transactions[:len(transactions)-1]
	}
	c.Data = transactions

	total, err := s.countTransactions(ctx, exec, q.Params)
	if err != nil {
		return c, err
	}
	c.Total = total

	return c, nil
}

func (s *Store) FindTransactions(ctx context.Context, q query.Query) (sharedapi.Cursor, error) {
	return s.findTransactions(ctx, s.schema, q)
}

func (s *Store) getTransaction(ctx context.Context, exec executor, txid uint64) (tx core.Transaction, err error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(
		"t.id",
		"t.timestamp",
		"t.reference",
		"t.metadata",
		"t.postings",
	)
	sb.From(sb.As(s.schema.Table("transactions"), "t"))
	sb.Where(sb.Equal("t.id", txid))
	sb.OrderBy("t.id DESC")

	sqlq, args := sb.BuildWithFlavor(s.schema.Flavor())
	rows, err := exec.QueryContext(ctx, sqlq, args...)
	if err != nil {
		return tx, s.error(err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			ref sql.NullString
			ts  sql.NullString
		)

		err := rows.Scan(
			&tx.ID,
			&ts,
			&ref,
			&tx.Metadata,
			&tx.Postings,
		)
		if err != nil {
			return tx, err
		}

		if tx.Metadata == nil {
			tx.Metadata = core.Metadata{}
		}
		t, err := time.Parse(time.RFC3339, ts.String)
		if err != nil {
			return tx, err
		}
		tx.Timestamp = t.UTC().Format(time.RFC3339)
		tx.Reference = ref.String
	}
	if rows.Err() != nil {
		return tx, s.error(rows.Err())
	}

	return tx, nil
}

func (s *Store) GetTransaction(ctx context.Context, txId uint64) (tx core.Transaction, err error) {
	return s.getTransaction(ctx, s.schema, txId)
}
