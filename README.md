# 1-Billion-Row Challenge (Go)

Welcome! This project is a simple but powerful example of how to process huge amounts of data (up to 1 billion rows) using Go. It is designed to be easy to understand, so you can learn about fast data processing, Go concurrency, and modern DevOps tools.

---

## 🚀 Features
- **Very fast data processing** using Go's goroutines (lightweight threads)
- **REST API** with Gin so you can send and get data easily
- **Modular code**: easy to read, change, and extend
- **Live monitoring**: see how your app is running with Prometheus & Grafana
- **Sample data and Postman collection** for easy testing
- **Runs in Docker**: start everything with one command

## 🗂️ Project Structure
- `src/` - Go source code
  - `main.go` - Program entry point
  - `delivery/` - Handles HTTP requests
  - `models/` - Data structures
  - `services/` - Main logic for processing data
  - `utilities/` - Helper functions
  - `test/` - Unit tests
- `assets/` - Example data and scripts
  - `sample/` - Example measurement files
  - `script/` - Load testing scripts
  - `postman_collection/` - Postman API collection
- `Dockerfile` - Build instructions for Docker
- `docker-compose.yml` - Runs Go app, Prometheus, and Grafana together
- `6671_rev2.json` - Example Grafana dashboard for Go metrics

---

## 🏁 Getting Started

### Prerequisites
- Install [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/)

### 🚦 Start Everything (App + Monitoring)
```sh
docker-compose up -d
```
- App: [http://localhost:8080](http://localhost:8080)
- Grafana: [http://localhost:3000](http://localhost:3000)  
  _Login: `admin` / `admin`_
- Prometheus: [http://localhost:9090](http://localhost:9090)

---

## 📈 Monitoring & Observability
- **Prometheus** collects Go app metrics automatically
- **Grafana** shows dashboards using Prometheus data
- **To use the sample dashboard:**
  1. Open Grafana ([localhost:3000](http://localhost:3000))
  2. Go to **Dashboards → Import**
  3. Upload `6671_rev2.json` from the project root
  4. Select Prometheus as the data source and click **Import**
- The dashboard shows:
  - Memory usage
  - Goroutine count
  - File descriptors
  - GC (garbage collection) duration

---

## 🧪 Running Tests
To run tests in Docker:
```sh
docker build --target tester -t 1brc-test .
docker run --rm 1brc-test
```
Or run locally:
```sh
cd src
go test ./...
```

---

## 📬 API Documentation
- Import the Postman collection from `assets/postman_collection/1-billion-row.postman_collection.json` into Postman to try the API endpoints.

---

## 📝 License
MIT

---

> _This project is great for learning about Go, fast data processing, and modern monitoring tools. Have fun exploring!_
