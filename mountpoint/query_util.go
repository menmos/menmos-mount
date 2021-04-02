package mountpoint

import (
	"github.com/menmos/menmos-go"
	"github.com/menmos/menmos-go/payload"
)

const queryBatchSize = 100

// aggregates all query results (using paging) into a single query response object.
func getFullQueryResults(query *payload.Query, client *menmos.Client) (*payload.QueryResponse, error) {
	response, err := client.Query(query.WithFrom(0).WithSize(queryBatchSize).WithSignURLs(false)) // No need to sign URLs.
	if err != nil {
		return nil, err
	}

	for response.Count < response.Total {
		resp, err := client.Query(query.WithFrom(response.Count))
		if err != nil {
			return nil, err
		}

		response.Count += resp.Count
		response.Hits = append(response.Hits, resp.Hits...)
	}

	return response, nil
}
