# data-pipe

A lightweight, standalone Change Data Capture (CDC) pipeline written in Go for syncing data between different databases without relying on Kafka or other external components.

## Features

- **Pure Go Implementation**: No external dependencies like Kafka required
- **MongoDB Change Streams**: Real-time CDC from MongoDB using native change streams
- **PostgreSQL Sink**: Efficient batch writes to PostgreSQL with upsert support
- **Field Mapping & Transformation**: Rename, format, and filter fields with powerful field mapper
- **Extensible Architecture**: Easy to add new sources (Convex, etc.) and sinks (ClickHouse, etc.)
- **Graceful Shutdown**: Properly handles SIGTERM and SIGINT signals
- **Configurable**: JSON-based configuration for easy setup
- **Batch Processing**: Optimized batch writes for better performance

## Architecture

The pipeline consists of three main components:

1. **Source**: Reads change events from a source database (MongoDB change streams)
2. **Transformer**: Optionally transforms events (pass-through by default)
3. **Sink**: Writes events to a target database (PostgreSQL with upsert logic)

### Pipeline Flow

```
MongoDB Change Stream → Events → [Transformer] → PostgreSQL
```

## Installation

### Prerequisites

- Go 1.20 or higher
- MongoDB 4.0+ (for change streams support)
- PostgreSQL 9.5+ (for upsert support)

### Build from Source

```bash
git clone https://github.com/IEatCodeDaily/data-pipe.git
cd data-pipe
go build -o data-pipe ./cmd/data-pipe
```

### Using Docker

Build and run with Docker:

```bash
# Build the Docker image
docker build -t data-pipe .

# Run with your config file
docker run -v $(pwd)/config.json:/root/config.json data-pipe
```

### Using Docker Compose

For a complete setup with MongoDB and PostgreSQL:

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f data-pipe

# Stop all services
docker-compose down
```

**Note**: When using docker-compose, you'll need to initialize the MongoDB replica set:

```bash
docker exec -it mongodb mongosh --eval "rs.initiate()"
```

## Configuration

Create a `config.json` file with your source and sink settings:

```json
{
  "pipeline": {
    "name": "mongodb-to-postgresql"
  },
  "source": {
    "type": "mongodb",
    "settings": {
      "uri": "mongodb://localhost:27017",
      "database": "mydb",
      "collection": "users"
    }
  },
  "sink": {
    "type": "postgresql",
    "settings": {
      "connection_string": "host=localhost port=5432 user=postgres password=postgres dbname=mydb sslmode=disable",
      "table": "users"
    }
  }
}
```

### Configuration Options

#### Pipeline Settings
- `name`: Identifier for the pipeline

#### MongoDB Source Settings
- `uri`: MongoDB connection string
- `database`: Database name to monitor
- `collection`: Collection name to monitor

#### PostgreSQL Sink Settings
- `connection_string`: PostgreSQL connection string
- `table`: Target table name

#### Transformer Settings (Optional)
- `type`: Transformer type (`passthrough` or `fieldmapper`)
- `settings`: Transformer-specific configuration

For detailed field mapping options, see [FIELD_MAPPING.md](FIELD_MAPPING.md).

**Field Mapping Example:**

```json
{
  "transformer": {
    "type": "fieldmapper",
    "settings": {
      "mappings": [
        {"source": "firstName", "destination": "first_name"},
        {"source": "email", "format": "lowercase"}
      ],
      "include_all": false
    }
  }
}
```

## Usage

### Running the Pipeline

```bash
./data-pipe -config config.json
```

### Command Line Options

- `-config`: Path to configuration file (default: "config.json")

### Example Workflow

1. **Prepare PostgreSQL Table**

   Create a table matching your MongoDB collection structure:

   ```sql
   CREATE TABLE users (
       _id TEXT PRIMARY KEY,
       name TEXT,
       email TEXT,
       created_at TIMESTAMP
   );
   ```

2. **Start the Pipeline**

   ```bash
   ./data-pipe -config config.json
   ```

3. **Test with MongoDB Operations**

   ```javascript
   // Insert a document in MongoDB
   db.users.insertOne({
     name: "John Doe",
     email: "john@example.com",
     created_at: new Date()
   });
   ```

   The data will automatically sync to PostgreSQL.

## Supported Operations

- **Insert**: Creates new records in PostgreSQL
- **Update**: Updates existing records (upsert)
- **Replace**: Replaces entire documents (upsert)
- **Delete**: Removes records from PostgreSQL

## How Sync Works

### Change Data Capture (CDC) Only

**Important**: The pipeline currently operates in **CDC-only mode**:

- ✅ Monitors for **new changes** after the pipeline starts
- ✅ Captures inserts, updates, and deletes in real-time
- ❌ Does **NOT** perform an initial full sync of existing data

**What this means:**
- Only documents created/modified/deleted **after** the pipeline starts are synced
- Pre-existing data in MongoDB is **not** automatically copied to PostgreSQL
- The pipeline listens to MongoDB's change stream (oplog) from the current point in time

### Initial Data Sync Strategies

For production deployments, you should perform an initial data load before starting the CDC pipeline:

#### Option 1: Manual Bulk Load

```bash
# Export from MongoDB
mongoexport --uri="mongodb://localhost:27017" --db=mydb --collection=users --out=users.json

# Import to PostgreSQL (custom script or tool)
# Then start the CDC pipeline
./data-pipe -config config.json
```

#### Option 2: MongoDB Tools

```bash
# Use mongodump/mongorestore or MongoDB Compass
mongodump --uri="mongodb://localhost:27017" --db=mydb --collection=users

# Convert and load to PostgreSQL, then start CDC
```

#### Option 3: Custom Script

Write a one-time script to:
1. Query all existing documents from MongoDB
2. Insert them into PostgreSQL
3. Start the data-pipe for ongoing CDC

### Resume from Specific Point

MongoDB change streams support resume tokens. To replay changes from a specific point in time:

```go
// Future enhancement - not currently implemented
opts := options.ChangeStream().SetStartAtOperationTime(&timestamp)
```

This feature is planned for future releases.

## Extending the Pipeline

### Adding a New Source

Implement the `Source` interface in `pkg/pipeline/types.go`:

```go
type Source interface {
    Connect(ctx context.Context) error
    Read(ctx context.Context) (<-chan Event, <-chan error)
    Close() error
}
```

Example structure for a Convex source:

```go
type ConvexSource struct {
    // connection details
}

func NewConvexSource(config ConvexConfig) *ConvexSource {
    // initialization
}

func (c *ConvexSource) Connect(ctx context.Context) error {
    // connect to Convex
}

func (c *ConvexSource) Read(ctx context.Context) (<-chan Event, <-chan error) {
    // read change events
}

func (c *ConvexSource) Close() error {
    // cleanup
}
```

### Adding a New Sink

Implement the `Sink` interface in `pkg/pipeline/types.go`:

```go
type Sink interface {
    Connect(ctx context.Context) error
    Write(ctx context.Context, events <-chan Event) <-chan error
    Close() error
}
```

Example structure for a ClickHouse sink:

```go
type ClickHouseSink struct {
    // connection details
}

func NewClickHouseSink(config ClickHouseConfig) *ClickHouseSink {
    // initialization
}

func (c *ClickHouseSink) Connect(ctx context.Context) error {
    // connect to ClickHouse
}

func (c *ClickHouseSink) Write(ctx context.Context, events <-chan Event) <-chan error {
    // write events in batches
}

func (c *ClickHouseSink) Close() error {
    // cleanup
}
```

## Project Structure

```
data-pipe/
├── cmd/
│   └── data-pipe/          # Main application entry point
│       └── main.go
├── pkg/
│   ├── pipeline/           # Core pipeline logic
│   │   ├── types.go        # Interfaces and types
│   │   └── pipeline.go     # Pipeline orchestration
│   ├── source/             # Source connectors
│   │   └── mongodb.go      # MongoDB source implementation
│   ├── sink/               # Sink connectors
│   │   └── postgresql.go   # PostgreSQL sink implementation
│   ├── transform/          # Data transformers
│   │   └── passthrough.go  # Pass-through transformer
│   └── config/             # Configuration management
│       └── config.go
├── examples/               # Example configurations
│   └── config.json
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
└── README.md
```

## Roadmap

- [x] MongoDB to PostgreSQL CDC
- [ ] Convex database source connector
- [ ] ClickHouse sink connector
- [ ] Custom data transformers
- [ ] Metrics and monitoring
- [ ] Multiple collection/table support
- [ ] Schema evolution handling
- [ ] State persistence and recovery

## Requirements

### MongoDB Requirements
- MongoDB 4.0+ for change stream support
- Replica set or sharded cluster (change streams don't work on standalone)

### PostgreSQL Requirements
- PostgreSQL 9.5+ for ON CONFLICT support
- Proper table schema matching MongoDB documents

## Performance Considerations

- **Batch Size**: Default batch size is 100 events. Adjust in `PostgreSQLSink` for your workload
- **Connection Pooling**: Both MongoDB and PostgreSQL drivers handle connection pooling
- **Network**: Keep source and sink close for lower latency

## Troubleshooting

### MongoDB Connection Issues

Ensure MongoDB is running as a replica set:
```bash
# Start MongoDB with replica set
mongod --replSet rs0

# Initialize replica set
mongosh --eval "rs.initiate()"
```

### PostgreSQL Connection Issues

Verify connection string format:
```
host=localhost port=5432 user=username password=password dbname=database sslmode=disable
```

### Change Stream Not Working

- Check MongoDB version (requires 4.0+)
- Verify replica set configuration
- Ensure proper authentication and permissions

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.