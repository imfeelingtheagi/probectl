// SPDX-License-Identifier: LicenseRef-probectl-TBD

package chclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

func urlEscape(s string) string { return url.QueryEscape(s) }

// Decode parses a ClickHouse JSONEachRow response body into row maps.
func Decode(body []byte) ([]map[string]any, error) {
	var rows []map[string]any
	for _, line := range bytes.Split(body, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var row map[string]any
		if err := json.Unmarshal(line, &row); err != nil {
			return nil, fmt.Errorf("chclient: decode row: %w", err)
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// Float coerces a JSONEachRow cell (number or numeric string) to float64.
func Float(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case string:
		f, _ := strconv.ParseFloat(n, 64)
		return f
	}
	return 0
}

// String coerces a cell to string ("" when not a string).
func String(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// Int coerces a cell to int (via Float).
func Int(v any) int { return int(Float(v)) }

// UintSlice coerces a ClickHouse array cell to []uint32.
func UintSlice(v any) []uint32 {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]uint32, 0, len(arr))
	for _, e := range arr {
		out = append(out, uint32(Float(e)))
	}
	return out
}

// Count extracts a single count() result keyed "n".
func Count(rows []map[string]any) int {
	if len(rows) == 0 {
		return 0
	}
	return Int(rows[0]["n"])
}
