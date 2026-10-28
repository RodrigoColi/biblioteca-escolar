package sqlutil

import (
	"database/sql"
	"strconv"
	"strings"
	"time"
)

// The edital forbids ORMs/query builders: statements are built by concatenating values,
// so text values are escaped here before going into the SQL string.
func Text(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func Num(n int) string {
	return strconv.Itoa(n)
}

func Date(t time.Time) string {
	return "'" + t.Format("2006-01-02") + "'"
}

func NullableDate(t *time.Time) string {
	if t == nil {
		return "NULL"
	}
	return Date(*t)
}

func TimePtr(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func Count(db *sql.DB, query string) (int, error) {
	var n int
	err := db.QueryRow(query).Scan(&n)
	return n, err
}
