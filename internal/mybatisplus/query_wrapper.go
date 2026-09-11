package mybatisplus

import (
	"fmt"
	"strings"
)

// The SQL fragments MyBatis-Plus's AbstractWrapper emits. The original never
// wrote them out because the library assembled the statement, so they are named
// here rather than returned as bare literals.
const (
	whereKeyword   = " WHERE "
	andSeparator   = " AND "
	orderByKeyword = " ORDER BY "
)

// QueryWrapper mirrors com.baomidou.mybatisplus.core.conditions.query.QueryWrapper
// for the two operations vueblog uses: eq(column, value) and orderByDesc(column).
type QueryWrapper struct {
	eqColumns []string
	eqValues  []any
	orderBys  []string
}

func NewQueryWrapper() *QueryWrapper { return &QueryWrapper{} }

// Eq mirrors QueryWrapper.eq(String column, Object val).
func (q *QueryWrapper) Eq(column string, value any) *QueryWrapper {
	q.eqColumns = append(q.eqColumns, column)
	q.eqValues = append(q.eqValues, value)
	return q
}

// OrderByDesc mirrors QueryWrapper.orderByDesc(String... columns).
func (q *QueryWrapper) OrderByDesc(columns ...string) *QueryWrapper {
	for _, c := range columns {
		q.orderBys = append(q.orderBys, c+" DESC")
	}
	return q
}

// WhereSQL renders the WHERE fragment and its bind arguments. Column names come
// from source literals, never user input, matching MP's own contract.
func (q *QueryWrapper) WhereSQL() (string, []any) {
	if len(q.eqColumns) == 0 {
		return "", nil
	}
	parts := make([]string, len(q.eqColumns))
	for i, c := range q.eqColumns {
		parts[i] = fmt.Sprintf("%s = ?", c)
	}
	return whereKeyword + strings.Join(parts, andSeparator), q.eqValues
}

// OrderSQL renders the ORDER BY fragment.
func (q *QueryWrapper) OrderSQL() string {
	if len(q.orderBys) == 0 {
		return ""
	}
	return orderByKeyword + strings.Join(q.orderBys, ", ")
}

// Eqs exposes the equality predicates for the in-memory mapper used by tests.
func (q *QueryWrapper) Eqs() ([]string, []any) { return q.eqColumns, q.eqValues }

// Orders exposes the ordering for the in-memory mapper.
func (q *QueryWrapper) Orders() []string { return q.orderBys }
