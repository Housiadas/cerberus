// Package elasticsearch is a client wrapper for Elasticsearch
// with OpenTelemetry instrumentation and bulk indexing support.
package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/telemetry"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"go.opentelemetry.io/otel/attribute"
)

// Document represents a document to be indexed or deleted in Elasticsearch.
type Document struct {
	Action string // "index" or "delete"
	ID     string
	Body   []byte // ignored for delete actions
}

// Client wraps the Elasticsearch client with logging and tracing.
type Client struct {
	es  *elasticsearch.Client
	log logger.Logger
}

// New creates a new Elasticsearch client.
func New(log logger.Logger, cfg Config) (*Client, error) {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 10,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConnectionFailed, err)
	}

	return &Client{
		es:  es,
		log: log,
	}, nil
}

// Ping checks if the Elasticsearch cluster is reachable.
func (c *Client) Ping(ctx context.Context) error {
	telemetry.AddSpan(ctx, "elasticsearch.Ping")

	res, err := c.es.Ping(c.es.Ping.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrConnectionFailed, err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("%w: status %s", ErrConnectionFailed, res.Status())
	}

	return nil
}

// CreateIndex creates an index with the given mapping if it does not already exist.
func (c *Client) CreateIndex(ctx context.Context, name string, mapping string) error {
	telemetry.AddSpan(ctx, "elasticsearch.CreateIndex",
		attribute.String("index", name),
	)

	// Check if index exists.
	res, err := c.es.Indices.Exists(
		[]string{name},
		c.es.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("checking index existence: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		c.log.Info(ctx, "elasticsearch", "status", "index already exists", "index", name)
		return nil
	}

	// Create the index.
	res, err = c.es.Indices.Create(
		name,
		c.es.Indices.Create.WithBody(strings.NewReader(mapping)),
		c.es.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("creating index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("creating index %s: %s %s", name, res.Status(), string(body))
	}

	c.log.Info(ctx, "elasticsearch", "status", "index created", "index", name)

	return nil
}

// BulkIndex indexes a batch of documents using the bulk API.
func (c *Client) BulkIndex(ctx context.Context, indexName string, docs []Document) error {
	telemetry.AddSpan(ctx, "elasticsearch.BulkIndex",
		attribute.String("index", indexName),
		attribute.Int("documents", len(docs)),
	)

	if len(docs) == 0 {
		return nil
	}

	indexer, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client: c.es,
		Index:  indexName,
	})
	if err != nil {
		return fmt.Errorf("%w: creating bulk indexer: %w", ErrBulkIndexFailed, err)
	}

	for _, doc := range docs {
		item := esutil.BulkIndexerItem{
			Action:     doc.Action,
			DocumentID: doc.ID,
		}
		if doc.Action != "delete" {
			item.Body = bytes.NewReader(doc.Body)
		}

		if err := indexer.Add(ctx, item); err != nil {
			return fmt.Errorf("%w: adding document %s: %w", ErrBulkIndexFailed, doc.ID, err)
		}
	}

	err = indexer.Close(ctx)
	if err != nil {
		return fmt.Errorf("%w: closing bulk indexer: %w", ErrBulkIndexFailed, err)
	}

	stats := indexer.Stats()

	c.log.Info(ctx, "elasticsearch",
		"status", "bulk index completed",
		"index", indexName,
		"indexed", stats.NumIndexed,
		"failed", stats.NumFailed,
	)

	if stats.NumFailed > 0 {
		return fmt.Errorf("%w: %d documents failed", ErrBulkIndexFailed, stats.NumFailed)
	}

	return nil
}

// DeleteDocument removes a document from the given index.
func (c *Client) DeleteDocument(ctx context.Context, indexName string, docID string) error {
	telemetry.AddSpan(ctx, "elasticsearch.DeleteDocument",
		attribute.String("index", indexName),
		attribute.String("documentId", docID),
	)

	res, err := c.es.Delete(
		indexName,
		docID,
		c.es.Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("deleting document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		// 404 is acceptable — the document may already be deleted.
		if res.StatusCode == http.StatusNotFound {
			return nil
		}

		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("deleting document %s: %s %s", docID, res.Status(), string(body))
	}

	return nil
}

// Search executes a search query against the given index.
func (c *Client) Search(ctx context.Context, indexName string, query map[string]any) (json.RawMessage, error) {
	telemetry.AddSpan(ctx, "elasticsearch.Search",
		attribute.String("index", indexName),
	)

	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("marshaling query: %w", err)
	}

	res, err := c.es.Search(
		c.es.Search.WithContext(ctx),
		c.es.Search.WithIndex(indexName),
		c.es.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, fmt.Errorf("executing search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		respBody, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search error: %s %s", res.Status(), string(respBody))
	}

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading search response: %w", err)
	}

	return respBody, nil
}
