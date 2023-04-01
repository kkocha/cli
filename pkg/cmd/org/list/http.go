package list

import (
	"context"
	"net/http"

	"github.com/cli/cli/v2/api"
	"github.com/shurcooL/githubv4"
)

type Organization struct {
	Login string
}

func listOrganizations(httpClient *http.Client, hostname string, limit int) ([]Organization, error) {
	type responseData struct {
		Viewer struct {
			Organizations struct {
				Nodes    []Organization
				PageInfo struct {
					HasNextPage bool
					EndCursor   string
				}
			} `graphql:"organizations(first: $perPage, after: $endCursor)"`
		}
	}

	maxPerPage := 1
	perPage := min(limit, maxPerPage)
	variables := map[string]interface{}{
		"perPage":   githubv4.Int(perPage),
		"endCursor": (*githubv4.String)(nil),
	}

	client := api.NewClientFromHTTP(httpClient)
	ctx := context.Background()

	var organizations []Organization
loop:
	for {
		var query responseData
		err := client.QueryWithContext(ctx, hostname, "OrganizationSearch", &query, variables)
		if err != nil {
			return nil, err
		}

		for _, org := range query.Viewer.Organizations.Nodes {
			organizations = append(organizations, org)
			if len(organizations) == limit {
				break loop
			}
		}

		if !query.Viewer.Organizations.PageInfo.HasNextPage {
			break
		}
		variables["endCursor"] = githubv4.String(query.Viewer.Organizations.PageInfo.EndCursor)
	}

	return organizations, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
