package utilities

import (
	"1brc-challange/models"
	"fmt"
	"mime/multipart"
	"os"
	"sync"
)

// memoryThreshold is the threshold for determining whether to read the file in memory or stream it to disk.
const memoryThreshold = 10 << 20 // 10MB

// SplitAndDecodeMultipartFileSmart splits and decodes a multipart file into parts, using memory or disk based on file size.
// It returns a slice of maps containing the decoded results for each part.
// If the file is small enough, it processes in memory; otherwise, it streams to a temporary file on disk.
// It uses goroutines to decode each part concurrently.
// The parts parameter specifies the number of parts to split the file into, typically the number of CPU cores available.
func SplitAndDecodeMultipartFileSmart(
	file multipart.File,
	header *multipart.FileHeader,
	parts int,
) ([]map[string]models.TempStat, error) {
	// Check if the file size is small enough to process in memory
	if header.Size <= memoryThreshold {
		// Decode the entire file in memory
		workerResults := make([]map[string]models.TempStat, 1)
		workerResults[0] = make(map[string]models.TempStat)
		err := DecodeMultipartFilePart(file, workerResults[0])
		if err != nil {
			return nil, fmt.Errorf("failed to decode multipart file part: %w", err)
		}
		defer file.Close()
		return workerResults, nil
	} else {
		// Large file: stream to disk once
		tempFile, err := streamToTempFile(file)
		if err != nil {
			return nil, nil
		}
		// Split the file into parts
		partsList, err := splitInDisk(tempFile, parts)
		if err != nil {
			tempFile.Close()
			os.Remove(tempFile.Name())
			return nil, nil
		}

		// Start decode workers
		var wg sync.WaitGroup
		workerResults := make([]map[string]models.TempStat, parts)
		// Initialize worker results
		fmt.Fprintf(os.Stderr, "🧵 Starting %d decode workers...\n", len(partsList))
		for i, p := range partsList {
			wg.Add(1)
			workerResults[i] = make(map[string]models.TempStat)
			go func(i int, p models.Part) {
				defer wg.Done()
				err := DecodePart(tempFile.Name(), p.Offset, p.Size, workerResults[i])
				if err != nil {
					fmt.Fprintf(os.Stderr, "Worker %d error: %v\n", i, err)
				}
			}(i, p)
		}
		wg.Wait()
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()
		return workerResults, nil
	}
}

// WARNING : Currently not used, but can be used to decode a part of a multipart file in memory.
// Parts is number of parts to split the file into. Usually this is the number of CPU cores available.
func SplitMultipartFileSmart(file multipart.File, header *multipart.FileHeader, parts int) ([]models.Part, error) {
	if header.Size <= memoryThreshold {
		return splitInMemory(file, parts)
	} else {
		// Large file: stream to disk and use seek-based logic
		tempFile, err := streamToTempFile(file)
		if err != nil {
			return nil, err
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()

		return splitInDisk(tempFile, parts)
	}
}

// WARNING : Currently not used, but can be used to decode a part of a multipart file in memory.
// DecodeMultipartFileSmart processes a multipart.File with offset and size, using memory or disk based on file size.
func DecodeMultipartFileSmart(file multipart.File, header *multipart.FileHeader, offset, size int64, result map[string]models.TempStat) error {
	if header.Size <= memoryThreshold {
		// Small file: decode in memory
		return DecodeMultipartFilePart(file, result)
	} else {
		// Large file: stream to disk and use disk-based logic
		tempFile, err := streamToTempFile(file)
		if err != nil {
			return err
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()
		return DecodePart(tempFile.Name(), offset, size, result)
	}
}
