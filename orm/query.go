package orm

import (
	"fmt"
	"strings"
)

// Query is a model-aware query builder that uses Go generics.
type Query[T any] struct {
	model        T
	table        string
	columns      []string
	wheres       []whereClause
	orders       []orderClause
	limitValue   int
	offsetValue  int
	withRelations []string
	withCounts   []string
	globalScopes []ScopeFunc[T]
	localScopes  []ScopeFunc[T]
	withTrashed  bool
	onlyTrashed  bool
}

type whereClause struct {
	column   string
	operator string
	value    any
	boolean  string // "and" or "or"
	isIn     bool
	values   []any
	isNull   bool
	isNotNull bool
	raw      string
	isRaw    bool
}

type orderClause struct {
	column    string
	direction string
}

// NewQuery creates a new query builder for the given model type.
func NewQuery[T any]() *Query[T] {
	return &Query[T]{
		columns: []string{"*"},
		wheres:  make([]whereClause, 0),
		orders:  make([]orderClause, 0),
		withRelations: make([]string, 0),
		withCounts: make([]string, 0),
		globalScopes: make([]ScopeFunc[T], 0),
		localScopes: make([]ScopeFunc[T], 0),
	}
}

// Table sets the table name for the query.
func (q *Query[T]) Table(table string) *Query[T] {
	q.table = table
	return q
}

// Select sets the columns to retrieve.
func (q *Query[T]) Select(columns ...string) *Query[T] {
	q.columns = columns
	return q
}

// Where adds a basic WHERE clause.
func (q *Query[T]) Where(column string, operator string, value any) *Query[T] {
	q.wheres = append(q.wheres, whereClause{
		column:   column,
		operator: operator,
		value:    value,
		boolean:  "and",
	})
	return q
}

// OrWhere adds an OR WHERE clause.
func (q *Query[T]) OrWhere(column string, operator string, value any) *Query[T] {
	q.wheres = append(q.wheres, whereClause{
		column:   column,
		operator: operator,
		value:    value,
		boolean:  "or",
	})
	return q
}

// WhereIn adds a WHERE IN clause.
func (q *Query[T]) WhereIn(column string, values []any) *Query[T] {
	q.wheres = append(q.wheres, whereClause{
		column:  column,
		boolean: "and",
		isIn:    true,
		values:  values,
	})
	return q
}

// OrWhereIn adds an OR WHERE IN clause.
func (q *Query[T]) OrWhereIn(column string, values []any) *Query[T] {
	q.wheres = append(q.wheres, whereClause{
		column:  column,
		boolean: "or",
		isIn:    true,
		values:  values,
	})
	return q
}

// WhereNull adds a WHERE NULL clause.
func (q *Query[T]) WhereNull(column string) *Query[T] {
	q.wheres = append(q.wheres, whereClause{
		column:  column,
		boolean: "and",
		isNull:  true,
	})
	return q
}

// WhereNotNull adds a WHERE NOT NULL clause.
func (q *Query[T]) WhereNotNull(column string) *Query[T] {
	q.wheres = append(q.wheres, whereClause{
		column:    column,
		boolean:   "and",
		isNotNull: true,
	})
	return q
}

// WhereRaw adds a raw WHERE clause.
func (q *Query[T]) WhereRaw(sql string) *Query[T] {
	q.wheres = append(q.wheres, whereClause{
		boolean: "and",
		isRaw:   true,
		raw:     sql,
	})
	return q
}

// OrderBy adds an ORDER BY clause.
func (q *Query[T]) OrderBy(column string, direction string) *Query[T] {
	q.orders = append(q.orders, orderClause{
		column:    column,
		direction: direction,
	})
	return q
}

// Limit sets the LIMIT clause.
func (q *Query[T]) Limit(limit int) *Query[T] {
	q.limitValue = limit
	return q
}

// Offset sets the OFFSET clause.
func (q *Query[T]) Offset(offset int) *Query[T] {
	q.offsetValue = offset
	return q
}

// With marks relations for eager loading.
func (q *Query[T]) With(relations ...string) *Query[T] {
	q.withRelations = append(q.withRelations, relations...)
	return q
}

// WithCount marks relations for count loading.
func (q *Query[T]) WithCount(relations ...string) *Query[T] {
	q.withCounts = append(q.withCounts, relations...)
	return q
}

// Has adds a relationship existence constraint.
func (q *Query[T]) Has(relation string) *Query[T] {
	// TODO: Implement relationship existence query
	return q
}

// WhereHas adds a relationship existence constraint with a callback.
func (q *Query[T]) WhereHas(relation string, fn func(*Query[T])) *Query[T] {
	// TODO: Implement relationship existence query with callback
	return q
}

// DoesntHave adds a relationship non-existence constraint.
func (q *Query[T]) DoesntHave(relation string) *Query[T] {
	// TODO: Implement relationship non-existence query
	return q
}

// WithTrashed includes soft-deleted records in the query.
func (q *Query[T]) WithTrashed() *Query[T] {
	q.withTrashed = true
	return q
}

// OnlyTrashed retrieves only soft-deleted records.
func (q *Query[T]) OnlyTrashed() *Query[T] {
	q.onlyTrashed = true
	return q
}

// Scope applies a local scope to the query.
func (q *Query[T]) Scope(fn ScopeFunc[T]) *Query[T] {
	q.localScopes = append(q.localScopes, fn)
	return q
}

// ToSQL generates the SQL query and bindings.
func (q *Query[T]) ToSQL() (string, []any) {
	var sql strings.Builder
	var bindings []any

	// SELECT
	sql.WriteString("SELECT ")
	sql.WriteString(strings.Join(q.columns, ", "))

	// FROM
	if q.table != "" {
		sql.WriteString(" FROM ")
		sql.WriteString(q.table)
	}

	// WHERE
	if len(q.wheres) > 0 {
		sql.WriteString(" WHERE ")
		for i, where := range q.wheres {
			if i > 0 {
				sql.WriteString(" ")
				sql.WriteString(strings.ToUpper(where.boolean))
				sql.WriteString(" ")
			}

			if where.isRaw {
				sql.WriteString(where.raw)
			} else if where.isNull {
				sql.WriteString(where.column)
				sql.WriteString(" IS NULL")
			} else if where.isNotNull {
				sql.WriteString(where.column)
				sql.WriteString(" IS NOT NULL")
			} else if where.isIn {
				sql.WriteString(where.column)
				sql.WriteString(" IN (")
				placeholders := make([]string, len(where.values))
				for j := range where.values {
					placeholders[j] = "?"
					bindings = append(bindings, where.values[j])
				}
				sql.WriteString(strings.Join(placeholders, ", "))
				sql.WriteString(")")
			} else {
				sql.WriteString(where.column)
				sql.WriteString(" ")
				sql.WriteString(where.operator)
				sql.WriteString(" ?")
				bindings = append(bindings, where.value)
			}
		}
	}

	// ORDER BY
	if len(q.orders) > 0 {
		sql.WriteString(" ORDER BY ")
		orderParts := make([]string, len(q.orders))
		for i, order := range q.orders {
			orderParts[i] = fmt.Sprintf("%s %s", order.column, strings.ToUpper(order.direction))
		}
		sql.WriteString(strings.Join(orderParts, ", "))
	}

	// LIMIT
	if q.limitValue > 0 {
		sql.WriteString(fmt.Sprintf(" LIMIT %d", q.limitValue))
	}

	// OFFSET
	if q.offsetValue > 0 {
		sql.WriteString(fmt.Sprintf(" OFFSET %d", q.offsetValue))
	}

	return sql.String(), bindings
}

// All retrieves all records.
func (q *Query[T]) All() ([]T, error) {
	// TODO: Execute query and return results
	return nil, nil
}

// Find retrieves a record by its primary key.
func (q *Query[T]) Find(id any) (*T, error) {
	// TODO: Execute query
	return nil, nil
}

// FindOrFail retrieves a record by its primary key or returns an error.
func (q *Query[T]) FindOrFail(id any) (*T, error) {
	result, err := q.Find(id)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("record not found")
	}
	return result, nil
}

// First retrieves the first record matching the query.
func (q *Query[T]) First() (*T, error) {
	q.Limit(1)
	// TODO: Execute query
	return nil, nil
}

// FirstOrFail retrieves the first record or returns an error.
func (q *Query[T]) FirstOrFail() (*T, error) {
	result, err := q.First()
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("record not found")
	}
	return result, nil
}

// Get retrieves all records matching the query.
func (q *Query[T]) Get() ([]T, error) {
	return q.All()
}

// Create inserts a new record.
func (q *Query[T]) Create(data map[string]any) (*T, error) {
	// TODO: Execute insert query
	return nil, nil
}

// CreateMany inserts multiple records.
func (q *Query[T]) CreateMany(data []map[string]any) ([]T, error) {
	// TODO: Execute batch insert query
	return nil, nil
}

// Update updates records matching the query.
func (q *Query[T]) Update(data map[string]any) (int64, error) {
	// TODO: Execute update query
	return 0, nil
}

// Delete soft deletes records matching the query.
func (q *Query[T]) Delete() (int64, error) {
	// TODO: Execute delete query (soft or hard based on model)
	return 0, nil
}

// ForceDelete permanently deletes records matching the query.
func (q *Query[T]) ForceDelete() (int64, error) {
	// TODO: Execute hard delete query
	return 0, nil
}

// Count returns the count of records matching the query.
func (q *Query[T]) Count() (int64, error) {
	// TODO: Execute count query
	return 0, nil
}

// Sum returns the sum of a column.
func (q *Query[T]) Sum(column string) (float64, error) {
	// TODO: Execute sum query
	return 0, nil
}

// Avg returns the average of a column.
func (q *Query[T]) Avg(column string) (float64, error) {
	// TODO: Execute avg query
	return 0, nil
}

// Max returns the maximum value of a column.
func (q *Query[T]) Max(column string) (any, error) {
	// TODO: Execute max query
	return nil, nil
}

// Min returns the minimum value of a column.
func (q *Query[T]) Min(column string) (any, error) {
	// TODO: Execute min query
	return nil, nil
}

// Paginator represents a paginated result set.
type Paginator[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	PerPage    int   `json:"per_page"`
	Page       int   `json:"page"`
	LastPage   int   `json:"last_page"`
	From       int   `json:"from"`
	To         int   `json:"to"`
	HasMore    bool  `json:"has_more"`
}

// Paginate paginates the query results.
func (q *Query[T]) Paginate(perPage, page int) (*Paginator[T], error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 15
	}

	// Count total
	total, err := q.Count()
	if err != nil {
		return nil, err
	}

	// Calculate pagination
	offset := (page - 1) * perPage
	lastPage := int((total + int64(perPage) - 1) / int64(perPage))

	// Fetch data
	q.Limit(perPage).Offset(offset)
	data, err := q.Get()
	if err != nil {
		return nil, err
	}

	from := offset + 1
	to := offset + len(data)
	if total == 0 {
		from = 0
		to = 0
	}

	return &Paginator[T]{
		Data:     data,
		Total:    total,
		PerPage:  perPage,
		Page:     page,
		LastPage: lastPage,
		From:     from,
		To:       to,
		HasMore:  page < lastPage,
	}, nil
}

// Chunk processes records in chunks.
func (q *Query[T]) Chunk(size int, fn func([]T) bool) error {
	page := 1
	for {
		q.Limit(size).Offset((page - 1) * size)
		results, err := q.Get()
		if err != nil {
			return err
		}

		if len(results) == 0 {
			break
		}

		if !fn(results) {
			break
		}

		if len(results) < size {
			break
		}

		page++
	}
	return nil
}
