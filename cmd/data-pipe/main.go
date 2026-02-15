package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/IEatCodeDaily/data-pipe/pkg/config"
	"github.com/IEatCodeDaily/data-pipe/pkg/pipeline"
	"github.com/IEatCodeDaily/data-pipe/pkg/sink"
	"github.com/IEatCodeDaily/data-pipe/pkg/source"
	"github.com/IEatCodeDaily/data-pipe/pkg/transform"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	logger := log.New(os.Stdout, "[data-pipe] ", log.LstdFlags)

	// Load configuration
	cfg, err := config.LoadFromFile(*configPath)
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	logger.Printf("Loaded configuration for pipeline: %s", cfg.Pipeline.Name)

	// Create source
	var src pipeline.Source
	switch cfg.Source.Type {
	case "mongodb":
		uri := cfg.Source.GetString("uri")
		database := cfg.Source.GetString("database")
		collection := cfg.Source.GetString("collection")
		src = source.NewMongoDBSource(uri, database, collection, logger)
	default:
		logger.Fatalf("Unsupported source type: %s", cfg.Source.Type)
	}

	// Create sink
	var snk pipeline.Sink
	switch cfg.Sink.Type {
	case "postgresql":
		connStr := cfg.Sink.GetString("connection_string")
		table := cfg.Sink.GetString("table")
		snk = sink.NewPostgreSQLSink(connStr, table, logger)
	default:
		logger.Fatalf("Unsupported sink type: %s", cfg.Sink.Type)
	}

	// Create transformer
	var transformer pipeline.Transformer
	
	if cfg.Transformer.Type != "" {
		switch cfg.Transformer.Type {
		case "fieldmapper":
			// Parse field mapper configuration
			if _, ok := cfg.Transformer.Settings["mappings"]; !ok {
				logger.Fatalf("fieldmapper transformer requires 'mappings' configuration")
			}
			
			// Convert settings to JSON and parse into FieldMapperConfig
			settingsJSON, err := json.Marshal(cfg.Transformer.Settings)
			if err != nil {
				logger.Fatalf("Failed to marshal transformer settings: %v", err)
			}
			
			var fmConfig transform.FieldMapperConfig
			if err := json.Unmarshal(settingsJSON, &fmConfig); err != nil {
				logger.Fatalf("Failed to parse fieldmapper configuration: %v", err)
			}
			
			fm, err := transform.NewFieldMapper(fmConfig)
			if err != nil {
				logger.Fatalf("Failed to create field mapper: %v", err)
			}
			transformer = fm
		case "passthrough":
			transformer = transform.NewPassThroughTransformer()
		default:
			logger.Fatalf("Unsupported transformer type: %s", cfg.Transformer.Type)
		}
	} else {
		// Default to passthrough if no transformer configured
		transformer = transform.NewPassThroughTransformer()
	}

	// Create pipeline
	pipe := pipeline.New(cfg.Pipeline.Name, src, snk, transformer, logger)

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Println("Received shutdown signal, stopping pipeline...")
		cancel()
	}()

	// Run pipeline
	logger.Println("Starting data pipeline...")
	if err := pipe.Run(ctx); err != nil {
		logger.Fatalf("Pipeline error: %v", err)
	}

	logger.Println("Pipeline stopped")
	fmt.Println("Goodbye!")
}
