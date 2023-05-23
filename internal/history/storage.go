package history

import "github.com/dtan4/bqc/internal/bigquery"

type Storage interface {
	Close() error
	Append(result *bigquery.Result) error
	List() ([]*bigquery.Result, error)
}
