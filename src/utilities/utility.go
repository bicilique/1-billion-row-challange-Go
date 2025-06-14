package utilities

import (
	"1brc-challange/models"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"sort"
	"strconv"
	"sync"
)

// ReadFile reads a file line by line and sends each line to the provided channel.
func ReadFile(path string, out chan<- []byte) {
	defer close(out)

	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	const bufSize = 4 * 1024 * 1024
	buf := make([]byte, bufSize)
	var leftover []byte

	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}

		chunk := append(leftover, buf[:n]...)
		lines := bytes.Split(chunk, []byte{'\n'})
		leftover = lines[len(lines)-1] // simpan baris sisa
		for _, line := range lines[:len(lines)-1] {
			if len(line) > 0 {
				out <- line
			}
		}

		if err == io.EOF {
			break
		}
	}

	if len(leftover) > 0 {
		out <- leftover
	}
}

// ReadMultipartFile reads a multipart.File line by line and sends each line to the provided channel.
func ReadMultipartFile(file multipart.File, out chan<- []byte) {
	defer close(out)

	const bufSize = 4 * 1024 * 1024
	buf := make([]byte, bufSize)
	var leftover []byte

	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}

		chunk := append(leftover, buf[:n]...)
		lines := bytes.Split(chunk, []byte{'\n'})
		leftover = lines[len(lines)-1]
		for _, line := range lines[:len(lines)-1] {
			if len(line) > 0 {
				out <- line
			}
		}

		if err == io.EOF {
			break
		}
	}

	if len(leftover) > 0 {
		out <- leftover
	}
}

// SplitLines splits lines from the input channel into station and temperature parts,
func SplitLines(in <-chan []byte, out chan<- models.LineSplit) {
	defer close(out)
	for line := range in {
		parts := bytes.SplitN(line, []byte(";"), 2)
		if len(parts) != 2 {
			continue
		}
		out <- models.LineSplit{Station: parts[0], Temperature: parts[1]}
	}
}

// LineSplitter splits a line into station and temperature parts.
func LineSplitter(line []byte) (models.LineSplit, bool) {
	station, temp, isFound := bytes.Cut(line, []byte(";"))
	if !isFound {
		return models.LineSplit{}, false
	}
	return models.LineSplit{Station: station, Temperature: temp}, true
}

// DecodeTemp decodes a byte slice representing a temperature value into a float32.
func DecodeTemp(tempBytes []byte) (float32, error) {
	if len(tempBytes) < 3 {
		return 0, fmt.Errorf("invalid temperature format")
	}
	var negative bool
	i := 0
	if tempBytes[i] == '-' {
		negative = true
		i++
	}
	var intPart int32
	for ; i < len(tempBytes) && tempBytes[i] != '.'; i++ {
		intPart = intPart*10 + int32(tempBytes[i]-'0')
	}
	if i+1 >= len(tempBytes) || tempBytes[i] != '.' {
		return 0, fmt.Errorf("invalid decimal format")
	}
	fracPart := int32(tempBytes[i+1] - '0')
	temp := float32(intPart) + float32(fracPart)/10
	if negative {
		temp = -temp
	}
	return temp, nil
}

// MergeResults merges multiple maps of TempStat into a single map.
func MergeResults(input []map[string]models.TempStat) map[string]*models.TempStat {
	final := make(map[string]*models.TempStat)
	for _, part := range input {
		for station, stat := range part {
			if existing, ok := final[station]; ok {
				existing.Sum += stat.Sum
				existing.Count += stat.Count
				if stat.Min < existing.Min {
					existing.Min = stat.Min
				}
				if stat.Max > existing.Max {
					existing.Max = stat.Max
				}
			} else {
				s := stat
				final[station] = &s
			}
		}
	}

	return final
}

// DecodePart reads a part of the file and decodes temperature data into a map of TempStat.
func DecodePart(path string, offset, size int64, result map[string]models.TempStat) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}
	limit := io.LimitReader(f, size)

	const bufSize = 1024 * 1024
	buf := make([]byte, bufSize)
	var leftover []byte

	stationCache := make(map[string]string)

	for {
		n, err := limit.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		chunk := append(leftover, buf[:n]...)
		lines := bytes.Split(chunk, []byte{'\n'})
		leftover = lines[len(lines)-1]

		for _, line := range lines[:len(lines)-1] {
			entry, ok := LineSplitter(line)
			if !ok {
				continue
			}

			raw := entry.Station
			key := string(raw)

			station, ok := stationCache[key]
			if !ok {
				copied := make([]byte, len(raw))
				copy(copied, raw)
				station = string(copied)
				stationCache[key] = station
			}

			temp, err := DecodeTemp(entry.Temperature)
			if err != nil {
				continue
			}

			stat, exists := result[station]
			if !exists {
				result[station] = models.TempStat{
					Sum:   temp,
					Min:   temp,
					Max:   temp,
					Count: 1,
				}
			} else {
				stat.Sum += temp
				stat.Count++
				if temp > stat.Max {
					stat.Max = temp
				}
				if temp < stat.Min {
					stat.Min = temp
				}
				result[station] = stat
			}
		}

		if err == io.EOF {
			break
		}
	}

	// process leftover
	if len(leftover) > 0 {
		entry, ok := LineSplitter(leftover)
		if ok {
			raw := entry.Station
			key := string(raw)

			station, ok := stationCache[key]
			if !ok {
				copied := make([]byte, len(raw))
				copy(copied, raw)
				station = string(copied)
				stationCache[key] = station
			}

			temp, err := DecodeTemp(entry.Temperature)
			if err == nil {
				stat, exists := result[station]
				if !exists {
					result[station] = models.TempStat{
						Sum:   temp,
						Min:   temp,
						Max:   temp,
						Count: 1,
					}
				} else {
					stat.Sum += temp
					stat.Count++
					if temp > stat.Max {
						stat.Max = temp
					}
					if temp < stat.Min {
						stat.Min = temp
					}
					result[station] = stat
				}
			}
		}
	}

	return nil
}

// DecodeMultipartFilePart reads a multipart.File and decodes temperature data into a map of TempStat.
func DecodeMultipartFilePart(file multipart.File, result map[string]models.TempStat) error {
	const bufSize = 1024 * 1024
	buf := make([]byte, bufSize)
	var leftover []byte

	stationCache := make(map[string]string)

	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		chunk := append(leftover, buf[:n]...)
		lines := bytes.Split(chunk, []byte{'\n'})
		leftover = lines[len(lines)-1]

		for _, line := range lines[:len(lines)-1] {
			entry, ok := LineSplitter(line)
			if !ok {
				continue
			}

			raw := entry.Station
			key := string(raw)

			station, ok := stationCache[key]
			if !ok {
				copied := make([]byte, len(raw))
				copy(copied, raw)
				station = string(copied)
				stationCache[key] = station
			}

			temp, err := DecodeTemp(entry.Temperature)
			if err != nil {
				continue
			}

			stat, exists := result[station]
			if !exists {
				result[station] = models.TempStat{
					Sum:   temp,
					Min:   temp,
					Max:   temp,
					Count: 1,
				}
			} else {
				stat.Sum += temp
				stat.Count++
				if temp > stat.Max {
					stat.Max = temp
				}
				if temp < stat.Min {
					stat.Min = temp
				}
				result[station] = stat
			}
		}

		if err == io.EOF {
			break
		}
	}

	// process leftover
	if len(leftover) > 0 {
		entry, ok := LineSplitter(leftover)
		if ok {
			raw := entry.Station
			key := string(raw)

			station, ok := stationCache[key]
			if !ok {
				copied := make([]byte, len(raw))
				copy(copied, raw)
				station = string(copied)
				stationCache[key] = station
			}

			temp, err := DecodeTemp(entry.Temperature)
			if err == nil {
				stat, exists := result[station]
				if !exists {
					result[station] = models.TempStat{
						Sum:   temp,
						Min:   temp,
						Max:   temp,
						Count: 1,
					}
				} else {
					stat.Sum += temp
					stat.Count++
					if temp > stat.Max {
						stat.Max = temp
					}
					if temp < stat.Min {
						stat.Min = temp
					}
					result[station] = stat
				}
			}
		}
	}

	return nil
}

// Parts is number of parts to split the file into. Usually this is the number of CPU cores available.
func SplitFile(path string, parts int) ([]models.Part, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return splitInDisk(f, parts)
}

// splitInMemory splits a file in memory into parts based on line offsets.
func splitInMemory(file multipart.File, parts int) ([]models.Part, error) {
	var buf bytes.Buffer
	offsets := make([]int64, 0)

	scanner := bufio.NewScanner(io.TeeReader(file, &buf))
	var currentOffset int64
	for scanner.Scan() {
		line := scanner.Bytes()
		lineLen := int64(len(line)) + 1
		offsets = append(offsets, currentOffset)
		currentOffset += lineLen
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	total := int64(buf.Len())
	return splitOffsets(offsets, total, parts), nil
}

// splitInDisk splits a file on disk into parts based on line offsets
func splitInDisk(f *os.File, parts int) ([]models.Part, error) {
	const maxLineLength = 100
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	size := info.Size()
	chunk := size / int64(parts)

	buf := make([]byte, maxLineLength)
	result := make([]models.Part, 0, parts)

	var offset int64
	for i := 0; i < parts; i++ {
		if i == parts-1 {
			result = append(result, models.Part{
				Offset: offset,
				Size:   size - offset})
			break
		}
		seek := offset + chunk
		if seek > size {
			seek = size
		}
		_, err := f.Seek(seek, io.SeekStart)
		if err != nil {
			return nil, err
		}
		n, _ := f.Read(buf)
		pos := bytes.IndexByte(buf[:n], '\n')
		if pos < 0 {
			return nil, fmt.Errorf("could not find newline at part %d", i)
		}
		cut := seek + int64(pos) + 1
		result = append(result, models.Part{
			Offset: offset,
			Size:   cut - offset})
		offset = cut
	}
	return result, nil
}

// streamToTempFile streams a multipart.File to a temporary file and returns the file handle.
func streamToTempFile(file multipart.File) (*os.File, error) {
	tmp, err := os.CreateTemp("", "upload-*.tmp")
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(tmp, file); err != nil {
		tmp.Close()
		return nil, err
	}
	_, err = tmp.Seek(0, io.SeekStart)
	return tmp, err
}

// splitOffsets splits the offsets into parts based on the total size and number of parts.
func splitOffsets(offsets []int64, total int64, parts int) []models.Part {
	if parts <= 0 {
		return nil
	}
	result := make([]models.Part, 0, parts)
	lines := len(offsets)
	linesPerPart := lines / parts

	for i := 0; i < parts; i++ {
		start := i * linesPerPart
		end := start + linesPerPart
		if i == parts-1 {
			end = lines
		}
		if start >= lines {
			break
		}

		startOffset := offsets[start]
		var endOffset int64
		if end >= lines {
			endOffset = total
		} else {
			endOffset = offsets[end]
		}
		result = append(result, models.Part{
			Offset: startOffset,
			Size:   endOffset - startOffset,
		})
	}
	return result
}

func DetectAnomalies(
	in <-chan models.LineSplit,
	out chan<- models.Anomaly,
	lastTemps map[string]float32,
	mu *sync.Mutex,
	totalAnomalies *int32,
	spikeCount *int32,
) {
	for entry := range in {
		temp, err := strconv.ParseFloat(string(entry.Temperature), 32)
		if err != nil {
			continue
		}
		t := float32(temp)
		station := string(entry.Station)

		isAnomaly := false
		reason := ""

		// Rule 1: extreme temperature
		if t < -50 || t > 60 {
			isAnomaly = true
			reason = "extreme"
		}

		// Rule 2: sudden spike (Δ > 20°C)
		mu.Lock()
		prev, exists := lastTemps[station]
		if exists && !isAnomaly && abs(t-prev) > 20.0 {
			isAnomaly = true
			reason = "spike"
			*spikeCount++
		}
		lastTemps[station] = t
		if isAnomaly {
			*totalAnomalies++
		}
		mu.Unlock()

		if isAnomaly {
			out <- models.Anomaly{Station: station, Temp: t, Reason: reason}
		}
	}
}

// WriteCSV writes the aggregated temperature statistics to a CSV file.
func WriteCSV(filename string, stats map[string]*models.TempStat) error {
	type row struct {
		Station string
		Mean    float32
		Min     float32
		Max     float32
	}

	var rows []row
	for station, stat := range stats {
		rows = append(rows, row{
			Station: station,
			Mean:    stat.Sum / float32(stat.Count),
			Min:     stat.Min,
			Max:     stat.Max,
		})
	}

	// Urutkan berdasarkan nama station
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Station < rows[j].Station
	})

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, r := range rows {
		_, err := fmt.Fprintf(writer, "%s;%.2f;%.2f;%.2f\n", r.Station, r.Mean, r.Min, r.Max)
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}

func WriteAnomalies(path string, in <-chan models.Anomaly) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	writer.WriteString("Station,Temp,Reason\n")

	for a := range in {
		fmt.Fprintf(writer, "%s,%.1f,%s\n", a.Station, a.Temp, a.Reason)
	}
	writer.Flush()
}

func abs(f float32) float32 {
	if f < 0 {
		return -f
	}
	return f
}
