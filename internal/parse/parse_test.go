package parse

import (
	"reflect"
	"strings"
	"testing"
)

const sampleHTML = `<html><body>
<table id="a"><thead><tr><th>كد درس</th><th>نام درس</th></tr></thead>
<tbody>
<tr><td>1</td><td>مبانی</td></tr>
<tr><td>2</td><td></td></tr>
</tbody></table>
<table><thead><tr><th>A</th><th>B</th></tr></thead><tbody><tr><td>x</td><td>y</td></tr></tbody></table>
</body></html>`

func TestParseTables(t *testing.T) {
	got, err := ParseTables(strings.NewReader(sampleHTML))
	if err != nil {
		t.Fatalf("ParseTables() error = %v", err)
	}

	want := []Table{
		{
			Headers: []string{"كد درس", "نام درس"},
			Rows:    [][]string{{"1", "مبانی"}, {"2", ""}},
		},
		{
			Headers: []string{"A", "B"},
			Rows:    [][]string{{"x", "y"}},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseTables() = %+v, want %+v", got, want)
	}
}

func TestParseTables_NoTables(t *testing.T) {
	got, err := ParseTables(strings.NewReader("<html><body><p>hi</p></body></html>"))
	if err != nil {
		t.Fatalf("ParseTables() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ParseTables() = %v, want no tables", got)
	}
}

// Truncated cells: the source export cuts the visible text with "..." and
// keeps the full text in a descendant title attribute; the title must win.
func TestParseTables_TruncatedCellUsesTitle(t *testing.T) {
	html := `<html><body><table><thead><tr><th>زمانبندي</th></tr></thead><tbody>
<tr><td><div title="سه شنبه از 07:15 تا 09:40 چهارشنبه از 13:00 تا 14:45">سه شنبه از 07:15 تا 09:40 چهارشنبه از ...</div></td></tr>
</tbody></table></body></html>`

	got, err := ParseTables(strings.NewReader(html))
	if err != nil {
		t.Fatalf("ParseTables() error = %v", err)
	}
	want := "سه شنبه از 07:15 تا 09:40 چهارشنبه از 13:00 تا 14:45"
	if got[0].Rows[0][0] != want {
		t.Errorf("cell = %q, want full text from title %q", got[0].Rows[0][0], want)
	}
}

// A title attribute without a "..." suffix is not a truncation marker; the
// visible text must win.
func TestParseTables_TitleWithoutTruncation(t *testing.T) {
	html := `<html><body><table><thead><tr><th>A</th></tr></thead><tbody>
<tr><td><div title="full text">visible text</div></td></tr>
</tbody></table></body></html>`

	got, err := ParseTables(strings.NewReader(html))
	if err != nil {
		t.Fatalf("ParseTables() error = %v", err)
	}
	if got[0].Rows[0][0] != "visible text" {
		t.Errorf("cell = %q, want visible text %q", got[0].Rows[0][0], "visible text")
	}
}
