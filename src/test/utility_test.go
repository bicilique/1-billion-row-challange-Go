package test

import (
	"1brc-challange/models"
	"1brc-challange/utilities"
	"bytes"
	"os"
	"sync"
	"testing"
)

func TestDecodeTemp(t *testing.T) {
	tests := []struct {
		input    []byte
		expected float32
		wantErr  bool
	}{
		{[]byte("23.4"), 23.4, false},
		{[]byte("-12.7"), -12.7, false},
		{[]byte("0.0"), 0.0, false},
		{[]byte("99.9"), 99.9, false},
		{[]byte("-0.1"), -0.1, false},
		{[]byte("bad"), 0, true},
	}
	for _, tt := range tests {
		got, err := utilities.DecodeTemp(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("DecodeTemp(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
		}
		if !tt.wantErr && got != tt.expected {
			t.Errorf("DecodeTemp(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestLineSplitter(t *testing.T) {
	line := []byte("StationA;12.3")
	res, ok := utilities.LineSplitter(line)
	if !ok {
		t.Fatal("LineSplitter failed to split valid line")
	}
	if string(res.Station) != "StationA" || string(res.Temperature) != "12.3" {
		t.Errorf("LineSplitter got %v, want StationA/12.3", res)
	}
}

func TestMergeResults(t *testing.T) {
	m1 := map[string]models.TempStat{"A": {Sum: 10, Min: 10, Max: 10, Count: 1}}
	m2 := map[string]models.TempStat{"A": {Sum: 20, Min: 5, Max: 20, Count: 2}, "B": {Sum: 30, Min: 30, Max: 30, Count: 1}}
	merged := utilities.MergeResults([]map[string]models.TempStat{m1, m2})
	if len(merged) != 2 {
		t.Errorf("Expected 2 stations, got %d", len(merged))
	}
	if merged["A"].Sum != 30 || merged["A"].Min != 5 || merged["A"].Max != 20 || merged["A"].Count != 3 {
		t.Errorf("MergeResults for A failed: %+v", merged["A"])
	}
	if merged["B"].Sum != 30 || merged["B"].Min != 30 || merged["B"].Max != 30 || merged["B"].Count != 1 {
		t.Errorf("MergeResults for B failed: %+v", merged["B"])
	}
}

func TestDetectAnomalies(t *testing.T) {
	in := make(chan models.LineSplit, 3)
	out := make(chan models.Anomaly, 3)
	lastTemps := make(map[string]float32)
	var mu sync.Mutex
	var totalAnomalies, spikeCount int32

	in <- models.LineSplit{Station: []byte("A"), Temperature: []byte("10.0")}
	in <- models.LineSplit{Station: []byte("A"), Temperature: []byte("35.0")}  // spike
	in <- models.LineSplit{Station: []byte("A"), Temperature: []byte("-60.0")} // extreme
	close(in)
	go utilities.DetectAnomalies(in, out, lastTemps, &mu, &totalAnomalies, &spikeCount)
	var results []models.Anomaly
	for a := range out {
		results = append(results, a)
		if len(results) == 2 {
			break
		}
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 anomalies, got %d", len(results))
	}
}

func TestWriteCSVAndRead(t *testing.T) {
	stats := map[string]*models.TempStat{
		"A": {Sum: 30, Min: 10, Max: 20, Count: 2},
		"B": {Sum: 40, Min: 15, Max: 25, Count: 2},
	}
	filename := "test_output.csv"
	err := utilities.WriteCSV(filename, stats)
	if err != nil {
		t.Fatalf("WriteCSV error: %v", err)
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if !bytes.Contains(data, []byte("A;15.00;10.00;20.00")) {
		t.Error("CSV missing expected row for A")
	}
	_ = os.Remove(filename)
}
