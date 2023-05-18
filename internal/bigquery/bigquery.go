package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

type Client struct {
	api *bigquery.Client
}

func NewClient(ctx context.Context, projectID string) (*Client, error) {
	api, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("create BigQuery client: %w", err)
	}

	return &Client{
		api: api,
	}, nil
}

func (c *Client) Close() error {
	return c.api.Close()
}

type Result struct {
	Keys                []string
	Rows                []map[string]bigquery.Value
	TotalBytesProcessed int64
}

func (c *Client) RunQuery(ctx context.Context, query string) (*Result, error) {
	q := c.api.Query(query)

	j, err := q.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("run BigQuery job: %w", err)
	}

	it, err := j.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("read BigQuery job result: %w", err)
	}

	keys := []string{}

	for _, r := range it.Schema {
		keys = append(keys, r.Name)
	}

	rows := []map[string]bigquery.Value{}

	for {
		var r map[string]bigquery.Value

		if err := it.Next(&r); err != nil {
			if err == iterator.Done {
				break
			}

			return nil, fmt.Errorf("load result: %w", err)
		}

		rows = append(rows, r)
	}

	s := j.LastStatus()
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("get the latest status: %w", err)
	}

	return &Result{
		Keys:                keys,
		Rows:                rows,
		TotalBytesProcessed: s.Statistics.TotalBytesProcessed,
	}, nil
}

type DryRunResult struct {
	TotalBytesProcessed int64
}

func (c *Client) DryRunQuery(ctx context.Context, query string) (*DryRunResult, error) {
	q := c.api.Query(query)
	q.DryRun = true

	j, err := q.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("run BigQuery job: %w", err)
	}

	s := j.LastStatus()
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("get the latest status: %w", err)
	}

	return &DryRunResult{
		TotalBytesProcessed: s.Statistics.TotalBytesProcessed,
	}, nil
}
