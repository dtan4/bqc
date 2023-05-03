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
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("create BigQuery client: %w", err)
	}
	defer client.Close()

	return &Client{
		api: client,
	}, nil
}

func (c *Client) Close() error {
	return c.api.Close()
}

func (c *Client) RunQuery(ctx context.Context, query string) ([]string, []map[string]bigquery.Value, error) {
	q := c.api.Query(query)

	j, err := q.Run(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("run BigQuery job: %w", err)
	}

	it, err := j.Read(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("read BigQuery job result: %w", err)
	}

	keys := []string{}

	for _, r := range it.Schema {
		keys = append(keys, r.Name)
	}

	var r map[string]bigquery.Value

	rows := []map[string]bigquery.Value{}

	for {
		if err := it.Next(&r); err != nil {
			if err == iterator.Done {
				break
			}

			return nil, nil, fmt.Errorf("load result: %w", err)
		}

		rows = append(rows, r)
	}

	return keys, rows, nil
}
