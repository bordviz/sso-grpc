package query

import "strings"

func QueryToString(q string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(q, "\n", " "),
		"\t",
		"",
	)
}
