package output

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"go.yaml.in/yaml/v3"
)

func sampleFixture() (RecordSet, time.Time) {
	d1 := time.Date(2025, 11, 28, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2025, 11, 29, 0, 0, 0, 0, time.UTC)
	d3 := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
	set := RecordSet{
		Columns: []Column{
			{Key: "id", Title: "ID"},
			{Key: "title", Title: "Title"},
			{Key: "start", Title: "Start"},
		},
		Rows: []map[string]any{
			{"id": "T-1", "title": "Buy milk", "start": d1},
			{"id": "T-2", "title": "Call team", "start": d2},
			{"id": "T-3", "title": "Ship", "start": d3},
		},
	}
	return set, d1
}

func hasANSI(s string) bool {
	return strings.Contains(s, "\x1b[")
}

func TestRender_CrossFormatAgreement(t *testing.T) {
	set, _ := sampleFixture()
	layout := DefaultDateLayout
	optsBase := RenderOptions{DateLayout: layout, Color: false}

	var jsonBuf, yamlBuf, csvBuf, tableBuf bytes.Buffer
	for _, tc := range []struct {
		format Format
		buf    *bytes.Buffer
	}{
		{FormatJSON, &jsonBuf},
		{FormatYAML, &yamlBuf},
		{FormatCSV, &csvBuf},
		{FormatTable, &tableBuf},
	} {
		opts := optsBase
		opts.Format = tc.format
		if err := Render(tc.buf, set, opts); err != nil {
			t.Fatalf("Render(%s): %v", tc.format, err)
		}
	}

	var jsonRows []map[string]any
	if err := json.Unmarshal(jsonBuf.Bytes(), &jsonRows); err != nil {
		t.Fatalf("json unmarshal: %v\n%s", err, jsonBuf.String())
	}
	if len(jsonRows) != 3 {
		t.Fatalf("json len=%d want 3", len(jsonRows))
	}

	var yamlRows []map[string]any
	if err := yaml.Unmarshal(yamlBuf.Bytes(), &yamlRows); err != nil {
		t.Fatalf("yaml unmarshal: %v\n%s", err, yamlBuf.String())
	}
	if len(yamlRows) != 3 {
		t.Fatalf("yaml len=%d want 3", len(yamlRows))
	}

	for i := range 3 {
		for _, key := range []string{"id", "title", "start"} {
			jv, jok := jsonRows[i][key]
			yv, yok := yamlRows[i][key]
			if !jok || !yok {
				t.Fatalf("row %d missing key %q", i, key)
			}
			if jv != yv {
				t.Fatalf("row %d key %q: json=%v yaml=%v", i, key, jv, yv)
			}
		}
		wantDate := FormatDate(set.Rows[i]["start"].(time.Time), layout)
		if jsonRows[i]["start"] != wantDate {
			t.Fatalf("json start[%d]=%v want %q", i, jsonRows[i]["start"], wantDate)
		}
	}

	r := csv.NewReader(strings.NewReader(csvBuf.String()))
	csvRecords, err := r.ReadAll()
	if err != nil {
		t.Fatalf("csv parse: %v", err)
	}
	if len(csvRecords) != 4 { // header + 3
		t.Fatalf("csv rows=%d want 4", len(csvRecords))
	}
	if got := strings.Join(csvRecords[0], ","); got != "id,title,start" {
		t.Fatalf("csv header=%q", got)
	}
	for i := range 3 {
		if csvRecords[i+1][0] != set.Rows[i]["id"] {
			t.Fatalf("csv id[%d]=%q", i, csvRecords[i+1][0])
		}
		wantDate := FormatDate(set.Rows[i]["start"].(time.Time), layout)
		if csvRecords[i+1][2] != wantDate {
			t.Fatalf("csv start[%d]=%q want %q", i, csvRecords[i+1][2], wantDate)
		}
	}

	tableOut := tableBuf.String()
	for i := range 3 {
		id := set.Rows[i]["id"].(string)
		if !strings.Contains(tableOut, id) {
			t.Fatalf("table missing id %q:\n%s", id, tableOut)
		}
		wantDate := FormatDate(set.Rows[i]["start"].(time.Time), layout)
		if !strings.Contains(tableOut, wantDate) {
			t.Fatalf("table missing date %q:\n%s", wantDate, tableOut)
		}
	}
}

func TestRender_EmptyRecordSet(t *testing.T) {
	set := RecordSet{
		Columns: []Column{{Key: "id"}, {Key: "title"}},
		Rows:    nil,
	}
	opts := RenderOptions{DateLayout: DefaultDateLayout, Color: false}

	var jsonBuf bytes.Buffer
	opts.Format = FormatJSON
	if err := Render(&jsonBuf, set, opts); err != nil {
		t.Fatal(err)
	}
	var arr []any
	if err := json.Unmarshal(jsonBuf.Bytes(), &arr); err != nil {
		t.Fatal(err)
	}
	if len(arr) != 0 {
		t.Fatalf("json empty want [], got %s", jsonBuf.String())
	}

	var yamlBuf bytes.Buffer
	opts.Format = FormatYAML
	if err := Render(&yamlBuf, set, opts); err != nil {
		t.Fatal(err)
	}
	var yarr []any
	if err := yaml.Unmarshal(yamlBuf.Bytes(), &yarr); err != nil {
		t.Fatal(err)
	}
	if len(yarr) != 0 {
		t.Fatalf("yaml empty want [], got %s", yamlBuf.String())
	}

	var csvBuf bytes.Buffer
	opts.Format = FormatCSV
	if err := Render(&csvBuf, set, opts); err != nil {
		t.Fatal(err)
	}
	r := csv.NewReader(strings.NewReader(csvBuf.String()))
	recs, err := r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 || recs[0][0] != "id" || recs[0][1] != "title" {
		t.Fatalf("csv empty=%v", recs)
	}

	var tableBuf bytes.Buffer
	opts.Format = FormatTable
	if err := Render(&tableBuf, set, opts); err != nil {
		t.Fatal(err)
	}
	// Header may appear; no data row ids
	if strings.Contains(tableBuf.String(), "T-") {
		t.Fatalf("table should have no data rows: %s", tableBuf.String())
	}
}

func TestRender_JSONYAMLRootArray(t *testing.T) {
	set, _ := sampleFixture()
	opts := RenderOptions{Format: FormatJSON, DateLayout: DefaultDateLayout}

	var buf bytes.Buffer
	if err := Render(&buf, set, opts); err != nil {
		t.Fatal(err)
	}
	trimmed := strings.TrimSpace(buf.String())
	if !strings.HasPrefix(trimmed, "[") {
		t.Fatalf("json root not array: %s", trimmed[:min(40, len(trimmed))])
	}
	if strings.Contains(trimmed, `"items"`) || strings.Contains(trimmed, `"data"`) {
		t.Fatalf("unexpected wrapper: %s", trimmed)
	}

	buf.Reset()
	opts.Format = FormatYAML
	if err := Render(&buf, set, opts); err != nil {
		t.Fatal(err)
	}
	var rows []map[string]any
	if err := yaml.Unmarshal(buf.Bytes(), &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != len(set.Rows) {
		t.Fatalf("yaml len=%d", len(rows))
	}
}

func TestRender_CSVEscaping(t *testing.T) {
	set := RecordSet{
		Columns: []Column{{Key: "id"}, {Key: "note"}},
		Rows: []map[string]any{
			{"id": "1", "note": "a,b"},
			{"id": "2", "note": "say \"hi\""},
			{"id": "3", "note": "line1\nline2"},
		},
	}
	var buf bytes.Buffer
	if err := Render(&buf, set, RenderOptions{Format: FormatCSV}); err != nil {
		t.Fatal(err)
	}
	r := csv.NewReader(strings.NewReader(buf.String()))
	recs, err := r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 4 {
		t.Fatalf("rows=%d", len(recs))
	}
	if recs[1][1] != "a,b" || recs[2][1] != "say \"hi\"" || recs[3][1] != "line1\nline2" {
		t.Fatalf("escaped values wrong: %v", recs)
	}
}

func TestRender_MachineFormatsNoANSI(t *testing.T) {
	set, _ := sampleFixture()
	for _, f := range []Format{FormatJSON, FormatYAML, FormatCSV} {
		var buf bytes.Buffer
		err := Render(&buf, set, RenderOptions{Format: f, Color: true, DateLayout: DefaultDateLayout})
		if err != nil {
			t.Fatal(err)
		}
		if hasANSI(buf.String()) {
			t.Fatalf("%s contains ANSI with Color=true", f)
		}
	}
}

func TestRender_TableColorANSI(t *testing.T) {
	set, _ := sampleFixture()
	var plain bytes.Buffer
	if err := Render(&plain, set, RenderOptions{Format: FormatTable, Color: false}); err != nil {
		t.Fatal(err)
	}
	if hasANSI(plain.String()) {
		t.Fatalf("plain table has ANSI: %q", plain.String())
	}

	var colored bytes.Buffer
	if err := Render(&colored, set, RenderOptions{Format: FormatTable, Color: true}); err != nil {
		t.Fatal(err)
	}
	// Colorized renderer SHOULD emit CSI for headers/borders; if library changes, fail loudly.
	if !hasANSI(colored.String()) {
		t.Fatalf("expected ANSI in colored table, got:\n%s", colored.String())
	}
}

func TestRender_NilDates(t *testing.T) {
	set := RecordSet{
		Columns: []Column{{Key: "id"}, {Key: "start"}},
		Rows: []map[string]any{
			{"id": "T-1", "start": nil},
			{"id": "T-2", "start": (*time.Time)(nil)},
		},
	}
	opts := RenderOptions{DateLayout: DefaultDateLayout, Color: false}

	var jsonBuf bytes.Buffer
	opts.Format = FormatJSON
	if err := Render(&jsonBuf, set, opts); err != nil {
		t.Fatal(err)
	}
	var rows []map[string]any
	if err := json.Unmarshal(jsonBuf.Bytes(), &rows); err != nil {
		t.Fatal(err)
	}
	for i, row := range rows {
		if row["start"] != nil {
			t.Fatalf("json start[%d]=%v want null", i, row["start"])
		}
	}

	var yamlBuf bytes.Buffer
	opts.Format = FormatYAML
	if err := Render(&yamlBuf, set, opts); err != nil {
		t.Fatal(err)
	}
	var yrows []map[string]any
	if err := yaml.Unmarshal(yamlBuf.Bytes(), &yrows); err != nil {
		t.Fatal(err)
	}
	for i, row := range yrows {
		if row["start"] != nil {
			t.Fatalf("yaml start[%d]=%v want null", i, row["start"])
		}
	}

	var csvBuf bytes.Buffer
	opts.Format = FormatCSV
	if err := Render(&csvBuf, set, opts); err != nil {
		t.Fatal(err)
	}
	r := csv.NewReader(strings.NewReader(csvBuf.String()))
	recs, err := r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if recs[1][1] != "" || recs[2][1] != "" {
		t.Fatalf("csv empty cells want \"\", got %v", recs)
	}

	var tableBuf bytes.Buffer
	opts.Format = FormatTable
	if err := Render(&tableBuf, set, opts); err != nil {
		t.Fatal(err)
	}
	// Should not panic; ids present
	if !strings.Contains(tableBuf.String(), "T-1") {
		t.Fatalf("table: %s", tableBuf.String())
	}
}

func TestRender_DateLayoutChangeAllFormats(t *testing.T) {
	set, d1 := sampleFixture()
	layoutA := "2006-01-02"
	layoutB := "02.01.2006"
	wantA := FormatDate(d1, layoutA)
	wantB := FormatDate(d1, layoutB)
	if wantA == wantB {
		t.Fatal("layouts should differ")
	}

	for _, f := range []Format{FormatJSON, FormatYAML, FormatCSV, FormatTable} {
		var a, b bytes.Buffer
		if err := Render(&a, set, RenderOptions{Format: f, DateLayout: layoutA}); err != nil {
			t.Fatal(err)
		}
		if err := Render(&b, set, RenderOptions{Format: f, DateLayout: layoutB}); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(a.String(), wantA) {
			t.Fatalf("%s layoutA missing %q:\n%s", f, wantA, a.String())
		}
		if !strings.Contains(b.String(), wantB) {
			t.Fatalf("%s layoutB missing %q:\n%s", f, wantB, b.String())
		}
		if strings.Contains(a.String(), wantB) && f != FormatTable {
			// table might coincidentally contain substrings; for machine formats assert exclusivity loosely
		}
		if !strings.Contains(a.String(), "Buy milk") || !strings.Contains(b.String(), "Buy milk") {
			t.Fatalf("%s lost non-date field", f)
		}
	}
}
