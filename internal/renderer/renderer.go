package renderer

import (
	"bytes"
	"encoding/csv"
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

type MarkdownRenderer struct{}

var _ Renderer = (*MarkdownRenderer)(nil)

func (r *MarkdownRenderer) Render(result *bigquery.Result) (string, error) {
	var b bytes.Buffer

	table := tablewriter.NewWriter(&b)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
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

type TSVRenderer struct{}

var _ Renderer = (*TSVRenderer)(nil)

func (r *TSVRenderer) Render(result *bigquery.Result) (string, error) {
	var b bytes.Buffer

	table := csv.NewWriter(&b)
	table.Comma = '\t'

	if err := table.Write(result.Keys); err != nil {
		return "", fmt.Errorf("write header to TSV: %w", err)
	}

	for _, r := range result.Rows {
		vs := []string{}
		for _, k := range result.Keys {
			vs = append(vs, fmt.Sprintf("%v", r[k]))
		}

		if err := table.Write(vs); err != nil {
			return "", fmt.Errorf("write row to TSV: %w", err)
		}
	}

	table.Flush()

	return b.String(), nil
}
