package services

import (
	"1brc-challange/models"
	"1brc-challange/utilities"
	"bytes"
	"fmt"
	"mime/multipart"
	"os"
	"runtime"
	"sync"
	"time"
)

type processService struct {
	NumCPU int
}

type ProcessService interface {
	OneBillionRowChallange(input multipart.File, header *multipart.FileHeader) (map[string]*models.TempStat, error)
	AnomalyDetection(input multipart.File) ([]*models.Anomaly, error)
}

func NewProcessService(numCPU int) ProcessService {
	fmt.Fprintf(os.Stderr, "🧠 CPU Cores Available : %d\n", numCPU)
	fmt.Fprintf(os.Stderr, "🧵 Decode Workers       : %d\n", numCPU)

	return &processService{
		NumCPU: numCPU,
	}
}

func (ps *processService) OneBillionRowChallange(input multipart.File, header *multipart.FileHeader) (map[string]*models.TempStat, error) {
	start := time.Now()
	// Validate the number of CPU cores
	if ps.NumCPU <= 0 {
		return nil, fmt.Errorf("invalid number of CPU cores: %d", ps.NumCPU)
	}
	// Validate the input file
	if input == nil || header == nil {
		return nil, fmt.Errorf("input file or header is nil")
	}
	// Validate the file size
	if header.Size <= 0 {
		return nil, fmt.Errorf("input file is empty or has invalid size: %d", header.Size)
	}
	// Split and decode the multipart file
	workerResults, err := utilities.SplitAndDecodeMultipartFileSmart(input, header, ps.NumCPU)
	if err != nil {
		return nil, fmt.Errorf("failed to decode multipart file: %w", err)
	}

	// Merge + output
	finalResult := utilities.MergeResults(workerResults)
	totalDone := time.Since(start)
	var logBuf bytes.Buffer
	showUsage(totalDone, &logBuf)
	fmt.Print(logBuf.String())
	return finalResult, nil
}

func (ps *processService) AnomalyDetection(input multipart.File) ([]*models.Anomaly, error) {
	start := time.Now()
	lines := make(chan []byte, 10000)
	splits := make(chan models.LineSplit, 10000)
	anomalies := make(chan models.Anomaly, 1000)

	// Shard stationTemps and mutexes per worker to reduce contention
	stationTempsShards := make([]map[string]float32, ps.NumCPU)
	statsMuShards := make([]sync.Mutex, ps.NumCPU)
	var anomalyCount int32
	var spikeCount int32

	// Initialize shards and mutexes
	go utilities.ReadMultipartFile(input, lines)

	// Split lines into LineSplit entries
	go utilities.SplitLines(lines, splits)

	// Initialize shards and mutexes
	var wg sync.WaitGroup
	for i := 0; i < ps.NumCPU; i++ {
		stationTempsShards[i] = make(map[string]float32)
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for entry := range splits {
				// Simple sharding: assign by hash of station name
				shard := int(entry.Station[0]) % ps.NumCPU
				if shard != idx {
					continue
				}
				utilities.DetectAnomalies(
					makeSingleEntryChan(entry),
					anomalies,
					stationTempsShards[idx],
					&statsMuShards[idx],
					&anomalyCount,
					&spikeCount,
				)
			}
		}(i)
	}

	// Wait for all workers to finish processing
	go func() {
		wg.Wait()
		close(anomalies)
	}()

	// Wait for all anomalies to be processed
	var detectedAnomalies []*models.Anomaly
	for anomaly := range anomalies {
		detectedAnomalies = append(detectedAnomalies, &anomaly)
	}

	// Buffered logging to avoid log interleaving
	var logBuf bytes.Buffer
	logBuf.WriteString("✅ Anomaly Detection Complete\n")
	logBuf.WriteString(fmt.Sprintf("📊 Total Anomalies Detected : %d\n", anomalyCount))
	logBuf.WriteString(fmt.Sprintf("⚠️  Spike Anomalies         : %d\n", spikeCount))

	// Calculate and log the total time taken
	totalDone := time.Since(start)
	showUsage(totalDone, &logBuf)
	fmt.Print(logBuf.String())

	return detectedAnomalies, nil
}

// makeSingleEntryChan returns a channel that yields a single LineSplit entry and then closes.
func makeSingleEntryChan(entry models.LineSplit) <-chan models.LineSplit {
	ch := make(chan models.LineSplit, 1)
	ch <- entry
	close(ch)
	return ch
}

func showUsage(totalDone time.Duration, logBuf *bytes.Buffer) {
	// Log the total time taken for the operation
	logBuf.WriteString(fmt.Sprintf("✅ Total Time Elapsed         : %.3fs\n", totalDone.Seconds()))

	// Memory usage statistics
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	logBuf.WriteString("💾 Memory Usage:\n")
	logBuf.WriteString(fmt.Sprintf("   Alloc        = %.2f MB\n", float64(mem.Alloc)/(1024*1024)))
	logBuf.WriteString(fmt.Sprintf("   TotalAlloc   = %.2f MB\n", float64(mem.TotalAlloc)/(1024*1024)))
	logBuf.WriteString(fmt.Sprintf("   Sys          = %.2f MB\n", float64(mem.Sys)/(1024*1024)))
	logBuf.WriteString(fmt.Sprintf("   NumGC        = %v\n", mem.NumGC))
}
