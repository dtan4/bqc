package renderer

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"

	"github.com/dtan4/bqc/internal/bigquery"
)

type Renderer interface {
	Render(result *bigquery.Result) (string, error)
}

type TableRenderer struct{}

var _ Renderer = (*TableRenderer)(nil)

func (r *TableRenderer) Render(result *bigquery.Result) (string, error) {
	var b bytes.Buffer

	ss := tw.NewSymbolCustom("Table").
		WithRow("-").
		WithColumn("|").
		WithTopLeft("+").
		WithTopMid("+").
		WithTopRight("+").
		WithMidLeft("+").
		WithCenter("+").
		WithMidRight("+").
		WithBottomLeft("+").
		WithBottomMid("+").
		WithBottomRight("+")

	table := tablewriter.NewTable(
		&b,
		tablewriter.WithRenderer(
			renderer.NewBlueprint(tw.Rendition{
				Settings: tw.Settings{
					Separators: tw.Separators{
						BetweenColumns: tw.On,
						BetweenRows:    tw.On,
					},
				},
				Symbols: ss,
			}),
		),
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Formatting: tw.CellFormatting{
					AutoWrap: tw.WrapNone,
				},
			},
			Row: tw.CellConfig{
				Formatting: tw.CellFormatting{
					Alignment: tw.AlignRight,
				},
			},
		}),
	)
	table.Configure(func(c *tablewriter.Config) {
		c.Header.Formatting.AutoFormat = false
	})

	table.Header(result.Keys)

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

	ss := tw.NewSymbolCustom("Table").
		WithRow("-").
		WithColumn("|").
		WithTopLeft("|").
		WithTopMid("|").
		WithTopRight("|").
		WithMidLeft("|").
		WithCenter("|").
		WithMidRight("|").
		WithBottomLeft("|").
		WithBottomMid("|").
		WithBottomRight("|")

	table := tablewriter.NewTable(
		&b,
		tablewriter.WithRenderer(
			renderer.NewBlueprint(tw.Rendition{
				Settings: tw.Settings{
					Separators: tw.Separators{
						BetweenColumns: tw.On,
						BetweenRows:    tw.Off,
					},
				},
				Symbols: ss,
				Borders: tw.Border{
					Left:   tw.On,
					Top:    tw.Off,
					Right:  tw.On,
					Bottom: tw.Off,
				},
			}),
		),
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Formatting: tw.CellFormatting{
					AutoWrap:  tw.WrapNone,
					Alignment: tw.AlignCenter,
				},
			},
			Row: tw.CellConfig{
				Formatting: tw.CellFormatting{
					Alignment: tw.AlignRight,
				},
			},
		}),
	)
	table.Configure(func(c *tablewriter.Config) {
		c.Header.Formatting.AutoFormat = false
	})

	table.Header(result.Keys)

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
