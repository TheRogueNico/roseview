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
