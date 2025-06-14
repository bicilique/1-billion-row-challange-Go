# 1-Billion-Row Challenge (Go)

This project solves the 1 Billion Row Challenge using Golang and Goroutines for high-performance data processing. The goal is to efficiently process and analyze extremely large datasets (up to 1 billion rows) with a focus on speed and concurrency.

## Features
- High-performance data processing using Go's concurrency model
- REST API built with Gin for data ingestion and querying
- Modular code structure for easy maintenance and extension
- Includes sample datasets and Postman collection for testing

## Project Structure
- `src/` - Main Go source code
  - `main.go` - Application entry point
  - `delivery/` - HTTP and routing logic
  - `models/` - Data models
  - `services/` - Business logic and processing
  - `utilities/` - Helper utilities
  - `test/` - Unit tests
- `assets/` - Sample datasets and scripts
  - `sample/` - Example measurement files
  - `script/` - k6 Test scripts
  - `postman_collection/` - Postman API collection
- `Dockerfile` - Multi-stage build for production-ready container
- `docker-compose.yml` - Easy local deployment

## Getting Started

### Prerequisites
- Docker and Docker Compose installed

### Build and Run with Docker Compose
```sh
docker-compose up --build
```
The app will be available at [http://localhost:8080](http://localhost:8080).

### Running Tests
The Dockerfile includes a test stage. To run tests manually:
```sh
docker build --target tester -t 1brc-test .
docker run --rm 1brc-test
```
Or run locally:
```sh
cd src
go test ./...
```

## API Documentation
- Import the Postman collection from `assets/postman_collection/1-billion-row.postman_collection.json` to explore available endpoints.

## License
MIT
