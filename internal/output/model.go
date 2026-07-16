package output

import (
	"fmt"
	"time"
)

// Column defines one field of a RecordSet.
type Column struct {
	Key   string // Stable key in json/yaml/csv headers
	Title string // Table header; Key used when empty
}

// RecordSet is an ordered set of homogeneous records.
type RecordSet struct {
	Columns []Column
	Rows    []map[string]any
}

// RenderOptions controls Render behavior.
type RenderOptions struct {
	Format       Format
	Color        bool   // ANSI only for table when true
	DateLayout   string // Go reference layout; empty/invalid → DefaultDateLayout
	SingleObject bool   // json/yaml: encode exactly one row as object (not array)
}

func (c Column) headerTitle() string {
	if c.Title != "" {
		return c.Title
	}
	return c.Key
}

func layoutOf(opts RenderOptions) string {
	return NormalizeLayout(opts.DateLayout)
}

// cellJSON returns a JSON/YAML-native value (dates as formatted strings; nil stays nil).
func cellJSON(v any, layout string) any {
	switch x := v.(type) {
	case nil:
		return nil
	case time.Time:
		return FormatDate(x, layout)
	case *time.Time:
		if x == nil {
			return nil
		}
		return FormatDate(*x, layout)
	default:
		return v
	}
}

// cellString returns a table/csv cell string (nil → "").
func cellString(v any, layout string) string {
	switch x := v.(type) {
	case nil:
		return ""
	case time.Time:
		return FormatDate(x, layout)
	case *time.Time:
		if x == nil {
			return ""
		}
		return FormatDate(*x, layout)
	case string:
		return x
	case bool:
		return fmt.Sprintf("%t", x)
	case int:
		return fmt.Sprintf("%d", x)
	case int64:
		return fmt.Sprintf("%d", x)
	case float64:
		return fmt.Sprintf("%v", x)
	case fmt.Stringer:
		return x.String()
	default:
		return fmt.Sprint(x)
	}
}

func validateRecordSet(set RecordSet) error {
	seen := make(map[string]struct{}, len(set.Columns))
	for _, c := range set.Columns {
		if c.Key == "" {
			return fmt.Errorf("output: empty column key")
		}
		if _, ok := seen[c.Key]; ok {
			return fmt.Errorf("output: duplicate column key %q", c.Key)
		}
		seen[c.Key] = struct{}{}
	}
	return nil
}

// mapsForJSONYAML builds []map[string]any with all column keys present.
func mapsForJSONYAML(set RecordSet, layout string) []map[string]any {
	out := make([]map[string]any, 0, len(set.Rows))
	for _, row := range set.Rows {
		m := make(map[string]any, len(set.Columns))
		for _, c := range set.Columns {
			var v any
			if row != nil {
				v = row[c.Key]
			}
			m[c.Key] = cellJSON(v, layout)
		}
		out = append(out, m)
	}
	return out
}
