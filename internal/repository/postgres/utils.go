package postgres_repository

import (
	"fmt"
	"strings"
)

func jsonObject(primaryKeyField string, fields map[string]string) string {
	parts := make([]string, 0, len(fields)*2)
	for k, v := range fields {
		parts = append(parts, fmt.Sprintf("'%s'", k), v)
	}
	return fmt.Sprintf(
		`CASE WHEN %s IS NULL THEN NULL ELSE jsonb_build_object(%s) END`,
		primaryKeyField,
		strings.Join(parts, ", "),
	)
}

func jsonArray(condition string, fields map[string]string) string {
	parts := make([]string, 0, len(fields)*2)
	for k, v := range fields {
		parts = append(parts, fmt.Sprintf("'%s'", k), v)
	}

	return fmt.Sprintf(
		`COALESCE(
			jsonb_agg(
				jsonb_build_object(%s)
			) FILTER (WHERE %s),
			'[]'
		)`,
		strings.Join(parts, ", "),
		condition,
	)
}
