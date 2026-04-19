package command

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Housiadas/cerberus/internal/sdk/elasticsearch"
	"github.com/Housiadas/cerberus/internal/sdk/host"
	"github.com/Housiadas/cerberus/internal/stream/indexer"
	"github.com/Housiadas/cerberus/pkg/kafka"
	"github.com/Housiadas/cerberus/pkg/telemetry"
)

const (
	indexerWorkers   = 4
	indexerBatchSize = 100
	indexerGroupID   = "cerberus-indexer"
)

// IndexerRunner implements Runner for the elasticsearch-indexer command.
type IndexerRunner struct {
	cmd *Command
}

func NewIndexerRunner(cmd *Command) *IndexerRunner {
	return &IndexerRunner{cmd: cmd}
}

func (r *IndexerRunner) Name() string               { return ElasticSearchIndexer }
func (r *IndexerRunner) Description() string        { return "Start the Elasticsearch indexer process" }
func (r *IndexerRunner) SetupFlags(_ *flag.FlagSet) {}
func (r *IndexerRunner) Run() error                 { return r.cmd.indexer() }

// Indexer starts the Elasticsearch indexer process that consumes Kafka messages
// and indexes data into Elasticsearch. Each registered handler consumes its own
// topic with a dedicated Kafka consumer.
func (cmd *Command) indexer() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = telemetry.InjectTracing(ctx, cmd.tracer)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// -------------------------------------------------------------------------
	// Elasticsearch client
	// -------------------------------------------------------------------------
	esClient, err := elasticsearch.New(cmd.log, elasticsearch.Config{
		Addresses: host.ParseHosts(cmd.esConfig.Hosts),
		Username:  cmd.esConfig.Username,
		Password:  cmd.esConfig.Password,
	})
	if err != nil {
		return fmt.Errorf("creating elasticsearch client: %w", err)
	}

	if err = esClient.Ping(ctx); err != nil {
		return fmt.Errorf("elasticsearch ping: %w", err)
	}

	// -------------------------------------------------------------------------
	// Register handlers
	// -------------------------------------------------------------------------
	handlers := []indexer.StreamIndexHandler{
		indexer.NewUserHandler(cmd.log, esClient),
	}

	// -------------------------------------------------------------------------
	// Create indices
	// -------------------------------------------------------------------------
	for _, h := range handlers {
		if err = esClient.CreateIndex(ctx, h.IndexName(), h.IndexMapping()); err != nil {
			return fmt.Errorf("creating index %s: %w", h.IndexName(), err)
		}
	}

	// -------------------------------------------------------------------------
	// Start one consumer per handler
	// -------------------------------------------------------------------------
	var wg sync.WaitGroup
	for _, h := range handlers {
		consumer, cErr := kafka.NewConsumer(cmd.log, kafka.ConsumerConfig{
			Brokers:          cmd.kafka.Brokers,
			GroupID:          indexerGroupID,
			AddressFamily:    cmd.kafka.AddressFamily,
			SecurityProtocol: cmd.kafka.SecurityProtocol,
			SessionTimeout:   cmd.kafka.SessionTimeout,
			Workers:          indexerWorkers,
			BatchSize:        indexerBatchSize,
		})
		if cErr != nil {
			cancel()
			wg.Wait()
			return fmt.Errorf("creating Kafka consumer for topic %s: %w", h.Topic(), cErr)
		}

		err = consumer.Subscribe(h.Topic())
		if err != nil {
			consumer.Close()
			cancel()
			wg.Wait()
			return fmt.Errorf("subscribing to topic %s: %w", h.Topic(), err)
		}

		cmd.log.Info(ctx,
			"indexer",
			"status", "consumer started",
			"topic", h.Topic(),
			"index", h.IndexName(),
		)

		wg.Add(1)
		go func(c *kafka.Consumer, sh indexer.StreamIndexHandler) {
			defer func() {
				c.Close()
				wg.Done()
			}()

			cErr := c.Consume(ctx, sh.Handle, sh.Flush)
			if cErr != nil {
				cmd.log.Error(ctx,
					"indexer",
					"status", "consume error",
					"topic", sh.Topic(),
					"msg", cErr,
				)
			}
		}(consumer, h)
	}

	// -------------------------------------------------------------------------
	// Shutdown
	// -------------------------------------------------------------------------
	sig := <-shutdown
	cmd.log.Info(ctx, "shutdown", "status", "indexer stopping", "signal", sig)

	cancel()
	wg.Wait()

	cmd.log.Info(ctx, "shutdown", "status", "indexer stopped")

	return nil
}
