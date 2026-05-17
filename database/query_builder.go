package database

import (
	"errors"
	"fmt"
	"strings"
)

// QueryBuilder builds SQL queries.
type QueryBuilder struct {
	conn          *Connection
	tableName     string
	selectColumns []string
	whereClause   []whereClause
	joins         []joinClause
	groupByClause []string
	havingClause  []whereClause
	orderClause   []orderClause
	limitValue    int
	offsetValue   int
	unionQueries  []*QueryBuilder
	lockMode      string
	bindings      []any
}

type whereClause struct {
	column   string
	operator string
	value    any
	boolean  string // "and" or "or"
	isRaw    bool
	isNull   bool
	isIn     bool
	isBetween bool
	isExists bool
	subquery func(*QueryBuilder)
}

type joinClause struct {
	joinType string
	table    string
	first    string
	operator string
	second   string
}

type orderClause struct {
	column    string
	direction string
	isRaw     bool
}

// newQueryBuilder creates a new query builder.
func newQueryBuilder(conn *Connection, table string) *QueryBuilder {
	return &QueryBuilder{
		conn:      conn,
		tableName: table,
		bindings:  make([]any, 0),
	}
}

// Table sets the table to query from.
func (qb *QueryBuilder) Table(name string) *QueryBuilder {
	qb.tableName = name
	return qb
}

// Select specifies columns to select.
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.selectColumns = append(qb.selectColumns, columns...)
	return qb
}

// SelectRaw adds a raw select expression.
func (qb *QueryBuilder) SelectRaw(expr string) *QueryBuilder {
	qb.selectColumns = append(qb.selectColumns, expr)
	return qb
}

// Where adds a basic WHERE clause.
func (qb *QueryBuilder) Where(column string, operator string, value any) *QueryBuilder {
	qb.whereClause = append(qb.whereClause, whereClause{
		column:   column,
		operator: operator,
		value:    value,
		boolean:  "and",
	})
	return qb
}

// OrWhere adds an OR WHERE clause.
func (qb *QueryBuilder) OrWhere(column string, operator string, value any) *QueryBuilder {
	qb.whereClause = append(qb.whereClause, whereClause{
		column:   column,
		operator: operator,
		value:    value,
		boolean:  "or",
	})
	return qb
}

// WhereIn adds a WHERE IN clause.
func (qb *QueryBuilder) WhereIn(column string, values []any) *QueryBuilder {
	qb.whereClause = append(qb.whereClause, whereClause{
		column:  column,
		value:   values,
		boolean: "and",
		isIn:    true,
	})
	return qb
}

// WhereNotIn adds a WHERE NOT IN clause.
func (qb *QueryBuilder) WhereNotIn(column string, values []any) *QueryBuilder {
	qb.whereClause = append(qb.whereClause, whereClause{
		column:   column,
		operator: "NOT IN",
		value:    values,
		boolean:  "and",
		isIn:     true,
	})
	return qb
}

// WhereNull adds a WHERE IS NULL clause.
func (qb *QueryBuilder) WhereNull(column string) *QueryBuilder {
	qb.whereClause = append(qb.whereClause, whereClause{
		column:  column,
		boolean: "and",
		isNull:  true,
	})
	return qb
}

// WhereNotNull adds a WHERE IS NOT NULL clause.
func (qb *QueryBuilder) WhereNotNull(column string) *QueryBuilder {
	qb.whereClause = append(qb.whereClause, whereClause{
		column:   column,
		operator: "NOT NULL",
		boolean:  "and",
		isNull:   true,
	})
	return qb
}

// WhereBetween adds a WHERE BETWEEN clause.
func (qb *QueryBuilder) WhereBetween(column string, min, max any) *QueryBuilder {
	qb.whereClause = append(qb.whereClause, whereClause{
		column:    column,
		value:     []any{min, max},
		boolean:   "and",
		isBetween: true,
	})
	return qb
}

// WhereExists adds a WHERE EXISTS clause.
func (qb *QueryBuilder) WhereExists(subquery func(*QueryBuilder)) *QueryBuilder {
	qb.whereClause = append(qb.whereClause, whereClause{
		boolean:  "and",
		isExists: true,
		subquery: subquery,
	})
	return qb
}

// Join adds an INNER JOIN clause.
func (qb *QueryBuilder) Join(table, first, operator, second string) *QueryBuilder {
	qb.joins = append(qb.joins, joinClause{
		joinType: "INNER",
		table:    table,
		first:    first,
		operator: operator,
		second:   second,
	})
	return qb
}

// LeftJoin adds a LEFT JOIN clause.
func (qb *QueryBuilder) LeftJoin(table, first, operator, second string) *QueryBuilder {
	qb.joins = append(qb.joins, joinClause{
		joinType: "LEFT",
		table:    table,
		first:    first,
		operator: operator,
		second:   second,
	})
	return qb
}

// CrossJoin adds a CROSS JOIN clause.
func (qb *QueryBuilder) CrossJoin(table string) *QueryBuilder {
	qb.joins = append(qb.joins, joinClause{
		joinType: "CROSS",
		table:    table,
	})
	return qb
}

// GroupBy adds a GROUP BY clause.
func (qb *QueryBuilder) GroupBy(columns ...string) *QueryBuilder {
	qb.groupByClause = append(qb.groupByClause, columns...)
	return qb
}

// Having adds a HAVING clause.
func (qb *QueryBuilder) Having(column, operator string, value any) *QueryBuilder {
	qb.havingClause = append(qb.havingClause, whereClause{
		column:   column,
		operator: operator,
		value:    value,
		boolean:  "and",
	})
	return qb
}

// OrderBy adds an ORDER BY clause.
func (qb *QueryBuilder) OrderBy(column, direction string) *QueryBuilder {
	qb.orderClause = append(qb.orderClause, orderClause{
		column:    column,
		direction: direction,
	})
	return qb
}

// OrderByDesc adds an ORDER BY DESC clause.
func (qb *QueryBuilder) OrderByDesc(column string) *QueryBuilder {
	return qb.OrderBy(column, "DESC")
}

// OrderByRaw adds a raw ORDER BY expression.
func (qb *QueryBuilder) OrderByRaw(expr string) *QueryBuilder {
	qb.orderClause = append(qb.orderClause, orderClause{
		column: expr,
		isRaw:  true,
	})
	return qb
}

// Limit sets the LIMIT clause.
func (qb *QueryBuilder) Limit(n int) *QueryBuilder {
	qb.limitValue = n
	return qb
}

// Offset sets the OFFSET clause.
func (qb *QueryBuilder) Offset(n int) *QueryBuilder {
	qb.offsetValue = n
	return qb
}

// Union adds a UNION clause.
func (qb *QueryBuilder) Union(other *QueryBuilder) *QueryBuilder {
	qb.unionQueries = append(qb.unionQueries, other)
	return qb
}

// LockForUpdate adds a FOR UPDATE lock.
func (qb *QueryBuilder) LockForUpdate() *QueryBuilder {
	qb.lockMode = "FOR UPDATE"
	return qb
}

// SharedLock adds a LOCK IN SHARE MODE lock.
func (qb *QueryBuilder) SharedLock() *QueryBuilder {
	qb.lockMode = "LOCK IN SHARE MODE"
	return qb
}

// ToSQL builds the SQL query and returns it with bindings.
func (qb *QueryBuilder) ToSQL() (string, []any) {
	qb.bindings = make([]any, 0)

	// Build SELECT columns
	columns := "*"
	if len(qb.selectColumns) > 0 {
		columns = strings.Join(qb.selectColumns, ", ")
	}

	sql := fmt.Sprintf("SELECT %s FROM %s", columns, qb.tableName)

	// Build JOINs
	for _, join := range qb.joins {
		if join.joinType == "CROSS" {
			sql += fmt.Sprintf(" CROSS JOIN %s", join.table)
		} else {
			sql += fmt.Sprintf(" %s JOIN %s ON %s %s %s",
				join.joinType, join.table, join.first, join.operator, join.second)
		}
	}

	// Build WHERE
	if len(qb.whereClause) > 0 {
		sql += " WHERE " + qb.buildWhereClauses(qb.whereClause)
	}

	// Build GROUP BY
	if len(qb.groupByClause) > 0 {
		sql += " GROUP BY " + strings.Join(qb.groupByClause, ", ")
	}

	// Build HAVING
	if len(qb.havingClause) > 0 {
		sql += " HAVING " + qb.buildWhereClauses(qb.havingClause)
	}

	// Build ORDER BY
	if len(qb.orderClause) > 0 {
		orders := make([]string, len(qb.orderClause))
		for i, order := range qb.orderClause {
			if order.isRaw {
				orders[i] = order.column
			} else {
				orders[i] = fmt.Sprintf("%s %s", order.column, order.direction)
			}
		}
		sql += " ORDER BY " + strings.Join(orders, ", ")
	}

	// Build LIMIT and OFFSET
	if qb.limitValue > 0 {
		sql += fmt.Sprintf(" LIMIT %d", qb.limitValue)
	}
	if qb.offsetValue > 0 {
		sql += fmt.Sprintf(" OFFSET %d", qb.offsetValue)
	}

	// Build LOCK
	if qb.lockMode != "" {
		sql += " " + qb.lockMode
	}

	// Build UNION
	for _, union := range qb.unionQueries {
		unionSQL, unionBindings := union.ToSQL()
		sql += " UNION " + unionSQL
		qb.bindings = append(qb.bindings, unionBindings...)
	}

	return sql, qb.bindings
}

func (qb *QueryBuilder) buildWhereClauses(clauses []whereClause) string {
	parts := make([]string, len(clauses))

	for i, clause := range clauses {
		prefix := ""
		if i > 0 {
			prefix = strings.ToUpper(clause.boolean) + " "
		}

		if clause.isExists {
			subQB := newQueryBuilder(qb.conn, "")
			clause.subquery(subQB)
			subSQL, subBindings := subQB.ToSQL()
			qb.bindings = append(qb.bindings, subBindings...)
			parts[i] = prefix + "EXISTS (" + subSQL + ")"
		} else if clause.isNull {
			if clause.operator == "NOT NULL" {
				parts[i] = prefix + clause.column + " IS NOT NULL"
			} else {
				parts[i] = prefix + clause.column + " IS NULL"
			}
		} else if clause.isIn {
			values := clause.value.([]any)
			placeholders := make([]string, len(values))
			for j := range values {
				placeholders[j] = "?"
				qb.bindings = append(qb.bindings, values[j])
			}
			op := "IN"
			if clause.operator == "NOT IN" {
				op = "NOT IN"
			}
			parts[i] = fmt.Sprintf("%s%s %s (%s)", prefix, clause.column, op, strings.Join(placeholders, ", "))
		} else if clause.isBetween {
			values := clause.value.([]any)
			parts[i] = fmt.Sprintf("%s%s BETWEEN ? AND ?", prefix, clause.column)
			qb.bindings = append(qb.bindings, values[0], values[1])
		} else {
			parts[i] = fmt.Sprintf("%s%s %s ?", prefix, clause.column, clause.operator)
			qb.bindings = append(qb.bindings, clause.value)
		}
	}

	return strings.Join(parts, " ")
}

// Get executes the query and returns all rows.
func (qb *QueryBuilder) Get() ([]map[string]any, error) {
	sql, bindings := qb.ToSQL()
	return qb.conn.Select(sql, bindings...)
}

// First executes the query and returns the first row.
func (qb *QueryBuilder) First() (map[string]any, error) {
	qb.Limit(1)
	sql, bindings := qb.ToSQL()
	return qb.conn.SelectOne(sql, bindings...)
}

// Count returns the count of rows.
func (qb *QueryBuilder) Count() (int64, error) {
	// Save original select columns
	originalSelect := qb.selectColumns
	qb.selectColumns = []string{"COUNT(*) as count"}

	row, err := qb.First()

	// Restore original select
	qb.selectColumns = originalSelect

	if err != nil {
		return 0, err
	}

	count, ok := row["count"].(int64)
	if !ok {
		return 0, errors.New("count query did not return int64")
	}

	return count, nil
}

// Sum returns the sum of a column.
func (qb *QueryBuilder) Sum(column string) (float64, error) {
	return qb.aggregate("SUM", column)
}

// Avg returns the average of a column.
func (qb *QueryBuilder) Avg(column string) (float64, error) {
	return qb.aggregate("AVG", column)
}

// Max returns the maximum value of a column.
func (qb *QueryBuilder) Max(column string) (float64, error) {
	return qb.aggregate("MAX", column)
}

// Min returns the minimum value of a column.
func (qb *QueryBuilder) Min(column string) (float64, error) {
	return qb.aggregate("MIN", column)
}

func (qb *QueryBuilder) aggregate(fn, column string) (float64, error) {
	originalSelect := qb.selectColumns
	qb.selectColumns = []string{fmt.Sprintf("%s(%s) as aggregate", fn, column)}

	row, err := qb.First()

	qb.selectColumns = originalSelect

	if err != nil {
		return 0, err
	}

	val, ok := row["aggregate"].(float64)
	if !ok {
		return 0, errors.New("aggregate query did not return float64")
	}

	return val, nil
}

// Insert inserts a row and returns the result.
func (qb *QueryBuilder) Insert(data map[string]any) (int64, error) {
	if len(data) == 0 {
		return 0, errors.New("no data to insert")
	}

	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]any, 0, len(data))

	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, "?")
		values = append(values, val)
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		qb.tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	result, err := qb.conn.Insert(sql, values...)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// InsertMany inserts multiple rows.
func (qb *QueryBuilder) InsertMany(data []map[string]any) error {
	if len(data) == 0 {
		return errors.New("no data to insert")
	}

	// Get columns from first row
	columns := make([]string, 0)
	for col := range data[0] {
		columns = append(columns, col)
	}

	// Build values clauses
	valueClauses := make([]string, len(data))
	allValues := make([]any, 0)

	for i, row := range data {
		placeholders := make([]string, len(columns))
		for j, col := range columns {
			placeholders[j] = "?"
			allValues = append(allValues, row[col])
		}
		valueClauses[i] = "(" + strings.Join(placeholders, ", ") + ")"
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		qb.tableName,
		strings.Join(columns, ", "),
		strings.Join(valueClauses, ", "))

	_, err := qb.conn.Insert(sql, allValues...)
	return err
}

// Update updates rows and returns affected count.
func (qb *QueryBuilder) Update(data map[string]any) (int64, error) {
	if len(data) == 0 {
		return 0, errors.New("no data to update")
	}

	sets := make([]string, 0, len(data))
	values := make([]any, 0, len(data))

	for col, val := range data {
		sets = append(sets, fmt.Sprintf("%s = ?", col))
		values = append(values, val)
	}

	sql := fmt.Sprintf("UPDATE %s SET %s", qb.tableName, strings.Join(sets, ", "))

	if len(qb.whereClause) > 0 {
		sql += " WHERE " + qb.buildWhereClauses(qb.whereClause)
		values = append(values, qb.bindings...)
	}

	return qb.conn.Update(sql, values...)
}

// Delete deletes rows and returns affected count.
func (qb *QueryBuilder) Delete() (int64, error) {
	sql := fmt.Sprintf("DELETE FROM %s", qb.tableName)

	qb.bindings = make([]any, 0)

	if len(qb.whereClause) > 0 {
		sql += " WHERE " + qb.buildWhereClauses(qb.whereClause)
	}

	return qb.conn.Delete(sql, qb.bindings...)
}

// Increment increments a column's value.
func (qb *QueryBuilder) Increment(column string, amount int) (int64, error) {
	sql := fmt.Sprintf("UPDATE %s SET %s = %s + ?", qb.tableName, column, column)

	values := []any{amount}

	if len(qb.whereClause) > 0 {
		sql += " WHERE " + qb.buildWhereClauses(qb.whereClause)
		values = append(values, qb.bindings...)
	}

	return qb.conn.Update(sql, values...)
}

// Decrement decrements a column's value.
func (qb *QueryBuilder) Decrement(column string, amount int) (int64, error) {
	sql := fmt.Sprintf("UPDATE %s SET %s = %s - ?", qb.tableName, column, column)

	values := []any{amount}

	if len(qb.whereClause) > 0 {
		sql += " WHERE " + qb.buildWhereClauses(qb.whereClause)
		values = append(values, qb.bindings...)
	}

	return qb.conn.Update(sql, values...)
}

// Chunk processes results in chunks.
func (qb *QueryBuilder) Chunk(size int, fn func([]map[string]any) bool) error {
	page := 0

	for {
		qb.Limit(size).Offset(page * size)

		results, err := qb.Get()
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

// Paginate returns paginated results.
func (qb *QueryBuilder) Paginate(perPage, page int) (*Paginator, error) {
	// Get total count
	total, err := qb.Count()
	if err != nil {
		return nil, err
	}

	// Calculate last page
	lastPage := int(total) / perPage
	if int(total)%perPage != 0 {
		lastPage++
	}

	// Get items for current page
	qb.Limit(perPage).Offset((page - 1) * perPage)
	items, err := qb.Get()
	if err != nil {
		return nil, err
	}

	return &Paginator{
		Items:       items,
		Total:       total,
		PerPage:     perPage,
		CurrentPage: page,
		LastPage:    lastPage,
	}, nil
}
