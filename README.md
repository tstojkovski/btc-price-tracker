# Bitcoin Price Tracker

A Go service that streams real-time Bitcoin (BTC) price data in USD to clients using Server-Sent Events (SSE).

## Features

- Fetches BTC/USD price from CoinGecko API every 10 seconds
- Streams real-time price updates to connected clients via SSE
- Supports historical data retrieval with the `?since=TIMESTAMP` query parameter
- Simple web interface for visualizing price updates
- In-memory storage with configurable capacity
- Proper concurrency handling using Go's goroutines and channels
- Clean shutdown with context cancellation
- Additional, configurable store implementation (MongoDB with TTL) in a PR [https://github.com/tstojkovski/btc-price-tracker/pull/1], since it's out of scope

## Architecture

The application follows a clean, modular architecture:

```
├── cmd/              # Application entry points
│   └── server/       # Main server application
├── internal/         # Internal packages
│   ├── domain/       # Domain models
│   ├── service/      # Business logic services
│   └── store/        # Data storage implementations
├── static/           # Static web assets
├── Dockerfile        # Docker configuration
└── Makefile          # Makefile
```

### Components

- **PriceService**: Fetches BTC price data and manages the update pipeline
- **BroadcastService**: Manages client connections and broadcasts updates
- **MemoryStore**: Thread-safe in-memory storage with circular buffer implementation
- **Web UI**: Simple HTML/JS frontend for visualizing price data

## Getting Started

### Prerequisites

- Go 1.23
- Internet connection (for API access)

### Running locally

1. Clone the repository
2. Build and run the server:

```bash
make build
make run
```

3. Open your browser to `http://localhost:8082`

### Configuration

The application can be configured using environment variables:

- `STORE_SIZE`: Number of price updates to keep in memory (default: 100)

### Docker

To build and run with Docker:

```bash
docker-build
docker-run
```

## API Endpoints

### `GET /prices/stream`

Server-Sent Events endpoint that streams BTC price updates.

**Parameters:**
- `since` (optional): Unix timestamp to retrieve historical data from

**Response Format:**
```json
{
  "timestamp": 1712525476,
  "price": 69420.25
}
```

## Production Readiness Considerations

### Scaling to 10,000+ Concurrent Users

1. **Horizontal Scaling**:
   - Deploy multiple instances behind a load balancer
   - Use a message broker (like Redis, Kafka, or NATS) for cross-instance communication

2. **Optimize Memory Usage**:
   - Implement connection timeouts and client cleanup
   - Consider a more efficient data structure for time series data

3. **Caching Layer**:
   - Add Redis for distributed caching of price data
   - Implement data partitioning by time windows

### Reliability & Fault Tolerance

1. **Resilient Price Fetching**:
   - Add multiple data sources with fallback options
   - Implement retry mechanisms with exponential backoff
   - Circuit breaker pattern for API calls

2. **Graceful Degradation**:
   - Return cached data when live data is unavailable
   - Implement rate limiting for client connections

3. **Data Persistence**:
   - Use a database for longer-term storage (e.g., MongoDB, TimescaleDB)
   - Implement periodic snapshots of in-memory data

### Observability

1. **Metrics Collection**:
   - Track client connections, disconnections, and reconnections
   - Monitor memory usage, API response times, and error rates
   - Implement Prometheus metrics endpoint

2. **Logging**:
   - Structured logging with correlation IDs
   - Log client connection events and API errors
   - Different log levels for development and production

3. **Distributed Tracing**:
   - OpenTelemetry integration for request tracing
   - Track end-to-end latency across services

### [Workflow](/docs/workflow.md) 
