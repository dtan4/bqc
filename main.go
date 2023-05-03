package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"

	"github.com/dtan4/bqc/internal/bigquery"
)

func main() {
	if err := realMain(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func realMain(args []string) error {
	if len(args) < 1 {
		return errors.New("project ID is required")
	}
	projectID := args[0]

	ctx := context.Background()

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("create BigQuery client: %w", err)
	}
	defer client.Close()

	sc := bufio.NewScanner(os.Stdin)

	for sc.Scan() {
		keys, rows, err := client.RunQuery(ctx, sc.Text())
		if err != nil {
			return fmt.Errorf("run query: %w", err)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoFormatHeaders(false)
		table.SetHeader(keys)

		for _, r := range rows {
			vs := []string{}
			for _, k := range keys {
				vs = append(vs, fmt.Sprintf("%v", r[k]))
			}

			table.Append(vs)
		}

		table.Render()
	}

	return nil
}
