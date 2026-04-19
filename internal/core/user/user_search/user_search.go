// Package user_search provides Elasticsearch-backed user search functionality.
package user_search

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Housiadas/cerberus/internal/sdk/elasticsearch"
	"github.com/Housiadas/cerberus/pkg/logger"
)

// searcher defines the interface for Elasticsearch search operations.
type searcher interface {
	Search(
		ctx context.Context,
		indexName string,
		query map[string]any,
	) (json.RawMessage, error)
}

// SearchFilter holds the available fields a search can be filtered on.
type SearchFilter struct {
	Q                string
	Email            string
	Department       string
	Enabled          *bool
	StartCreatedDate *time.Time
	EndCreatedDate   *time.Time
}

// UserDoc represents a user document as stored in Elasticsearch.
type UserDoc struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Department string `json:"department"`
	Enabled    bool   `json:"enabled"`
	AccountID  string `json:"accountId"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

// SearchResult holds the search results and pagination metadata.
type SearchResult struct {
	Users       []UserDoc
	TotalHits   int
	HasMore     bool
	SearchAfter string
}

// Store manages Elasticsearch search operations for users.
type Store struct {
	es  searcher
	log logger.Logger
}

// NewStore constructs a new user search store.
func NewStore(log logger.Logger, es searcher) *Store {
	return &Store{
		log: log,
		es:  es,
	}
}

// SearchUsers executes a search query against the users Elasticsearch index.
func (s *Store) SearchUsers(
	ctx context.Context,
	filter SearchFilter,
	sort string,
	limit int,
	searchAfter string,
) (SearchResult, error) {
	query := buildQuery(filter, sort, limit, searchAfter)

	raw, err := s.es.Search(ctx, elasticsearch.UserIndex, query)
	if err != nil {
		return SearchResult{}, fmt.Errorf("user search: %w", err)
	}

	return parseResponse(raw, limit)
}

// -------------------------------------------------------------------------
// Query builder
// -------------------------------------------------------------------------

func buildQuery(
	filter SearchFilter,
	sort string,
	limit int,
	searchAfter string,
) map[string]any {
	query := make(map[string]any)

	// Bool query
	boolQuery := make(map[string]any)

	// Must clauses (full-text search)
	if filter.Q != "" {
		boolQuery["must"] = []map[string]any{
			{
				"match": map[string]any{
					"name": map[string]any{
						"query":     filter.Q,
						"fuzziness": "AUTO",
					},
				},
			},
		}
	}

	// Filter clauses (exact matches, ranges)
	var filters []map[string]any

	if filter.Email != "" {
		filters = append(filters, map[string]any{
			"term": map[string]any{"email": filter.Email},
		})
	}

	if filter.Department != "" {
		filters = append(filters, map[string]any{
			"term": map[string]any{"department": filter.Department},
		})
	}

	if filter.Enabled != nil {
		filters = append(filters, map[string]any{
			"term": map[string]any{"enabled": *filter.Enabled},
		})
	}

	if filter.StartCreatedDate != nil || filter.EndCreatedDate != nil {
		rangeFilter := make(map[string]any)
		if filter.StartCreatedDate != nil {
			rangeFilter["gte"] = filter.StartCreatedDate.UTC().Format(time.RFC3339)
		}
		if filter.EndCreatedDate != nil {
			rangeFilter["lte"] = filter.EndCreatedDate.UTC().Format(time.RFC3339)
		}
		filters = append(filters, map[string]any{
			"range": map[string]any{"createdAt": rangeFilter},
		})
	}

	if len(filters) > 0 {
		boolQuery["filter"] = filters
	}

	// Exclude soft-deleted users
	boolQuery["must_not"] = []map[string]any{
		{"exists": map[string]any{"field": "deletedAt"}},
	}

	query["query"] = map[string]any{"bool": boolQuery}

	// Sort
	sortField, sortDir := parseSortParam(sort, filter.Q != "")
	query["sort"] = []map[string]any{
		{sortField: sortDir},
		{"id": "asc"}, // tiebreaker for search_after
	}

	// Size: request limit+1 to determine hasMore
	query["size"] = limit + 1

	// search_after pagination
	if searchAfter != "" {
		decoded, err := base64.URLEncoding.DecodeString(searchAfter)
		if err == nil {
			var sa []any
			if json.Unmarshal(decoded, &sa) == nil {
				query["search_after"] = sa
			}
		}
	}

	return query
}

func parseSortParam(sort string, hasQuery bool) (string, string) {
	if sort == "" {
		if hasQuery {
			return "_score", "desc"
		}
		return "createdAt", "desc"
	}

	// Expected format: "field:direction"
	for i := len(sort) - 1; i >= 0; i-- {
		if sort[i] == ':' {
			field := sort[:i]
			dir := sort[i+1:]
			if dir != "asc" && dir != "desc" {
				dir = "desc"
			}
			return field, dir
		}
	}

	return sort, "desc"
}

// -------------------------------------------------------------------------
// Response parser
// -------------------------------------------------------------------------

type esResponse struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []esHit `json:"hits"`
	} `json:"hits"`
}

type esHit struct {
	Source UserDoc `json:"_source"`
	Sort   []any   `json:"sort"`
}

func parseResponse(raw json.RawMessage, limit int) (SearchResult, error) {
	var resp esResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return SearchResult{}, fmt.Errorf("parsing search response: %w", err)
	}

	hits := resp.Hits.Hits
	hasMore := len(hits) > limit
	if hasMore {
		hits = hits[:limit]
	}

	users := make([]UserDoc, 0, len(hits))
	for _, hit := range hits {
		users = append(users, hit.Source)
	}

	var searchAfter string
	if len(hits) > 0 {
		lastSort := hits[len(hits)-1].Sort
		sortBytes, err := json.Marshal(lastSort)
		if err == nil {
			searchAfter = base64.URLEncoding.EncodeToString(sortBytes)
		}
	}

	return SearchResult{
		Users:       users,
		TotalHits:   resp.Hits.Total.Value,
		HasMore:     hasMore,
		SearchAfter: searchAfter,
	}, nil
}
