package renderer

import (
	"bytes"
	"fmt"

	"github.com/olekukonko/tablewriter"

	"github.com/dtan4/bqc/internal/bigquery"
)

type Renderer interface {
	Render(result *bigquery.Result) (string, error)
}

type TableRenderer struct{}

var _ Renderer = (*TableRenderer)(nil)

func (r *TableRenderer) Render(result *bigquery.Result) (string, error) {
	var b bytes.Buffer

	table := tablewriter.NewWriter(&b)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.SetHeader(result.Keys)

	for _, r := range result.Rows {
		vs := []string{}
		for _, k := range result.Keys {
			vs = append(vs, fmt.Sprintf("%v", r[k]))
		}

		table.Append(vs)
	}

	table.Render()

	return b.String(), nil
}
