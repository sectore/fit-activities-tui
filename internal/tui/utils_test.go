package tui

import (
	"testing"
	"time"

	"github.com/sectore/fit-activities-tui/internal/common"
)

func TestHorizontalStackedBar(t *testing.T) {
	tests := []struct {
		name        string
		value1      float64
		value1Block string
		value2      float64
		value2Block string
		maxBlocks   int
		expected    string
	}{
		{
			name:        "active 42m 39s pause 3m 41s",
			value1:      2559, // active seconds
			value1Block: "▒",
			value2:      221, // pause seconds
			value2Block: "░",
			maxBlocks:   50,
			expected:    "▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒░░░░",
		},
		{
			name:        "active 1h 11m 45s pause 33m 47s",
			value1:      4305, // active seconds
			value1Block: "▒",
			value2:      2027, // pause seconds
			value2Block: "░",
			maxBlocks:   50,
			expected:    "▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒░░░░░░░░░░░░░░░░",
		},
		{
			name:        "active 28m 53s pause 30s",
			value1:      1733, // active seconds
			value1Block: "▒",
			value2:      30, // pause seconds
			value2Block: "░",
			maxBlocks:   50,
			expected:    "▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒░",
		},
		{
			name:        "active 2h 58m 13s pause 1h 51m 44s",
			value1:      10693, // active seconds
			value1Block: "▒",
			value2:      6704, // pause seconds
			value2Block: "░",
			maxBlocks:   50,
			expected:    "▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒░░░░░░░░░░░░░░░░░░░",
		},
		{
			name:        "1234 | 0",
			value1:      1234,
			value1Block: "▒",
			value2:      0,
			value2Block: "░",
			maxBlocks:   50,
			expected:    "▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒",
		},
		{
			name:        "disproportionate allocation due to truncation",
			value1:      50.0,
			value1Block: "▒",
			value2:      49.9, // Almost equal values, should get nearly equal blocks
			value2Block: "░",
			maxBlocks:   8,
			expected:    "▒▒▒▒░░░░", // Should be 4 and 4, but current gives 5 and 3 due to truncation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HorizontalStackedBar(tt.value1, tt.value1Block, tt.value2, tt.value2Block, tt.maxBlocks)

			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestTimeToDuration(t *testing.T) {
	tests := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected string
	}{
		{
			name:     "5 seconds",
			start:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 1, 1, 10, 0, 5, 0, time.UTC),
			expected: "5s",
		},
		{
			name:     "30 seconds",
			start:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 1, 1, 10, 0, 30, 0, time.UTC),
			expected: "30s",
		},
		{
			name:     "2 minutes 15 seconds",
			start:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 1, 1, 10, 2, 15, 0, time.UTC),
			expected: "2m 15s",
		},
		{
			name:     "1 hour 30 minutes 45 seconds",
			start:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 1, 1, 11, 30, 45, 0, time.UTC),
			expected: "1h 30m 45s",
		},
		{
			name:     "5 hours 0 minutes 0 seconds",
			start:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
			expected: "5h 0m 0s",
		},
		{
			name:     "2 days 5 hours 30 minutes 15 seconds",
			start:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 1, 3, 15, 30, 15, 0, time.UTC),
			expected: "2d 5h 30m 15s",
		},
		{
			name:     "3 days 0 hours 0 minutes 0 seconds",
			start:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 1, 4, 12, 0, 0, 0, time.UTC),
			expected: "3d 0h 0m 0s",
		},
		{
			name:     "zero duration (same time)",
			start:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			expected: "0s",
		},
		{
			name:     "start after end (negative duration should be zero)",
			start:    time.Date(2024, 1, 1, 15, 30, 0, 0, time.UTC),
			end:      time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			expected: "0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startTime := common.NewTime(tt.start)
			endTime := common.NewTime(tt.end)

			result := TimeToDuration(startTime, endTime)

			if result.Format() != tt.expected {
				t.Errorf("Expected: %s, Got: %s", tt.expected, result.Format())
			}
		})
	}
}

func TestHorizontalBar(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		fgBlock   string
		maxValue  float64
		bgBlock   string
		maxBlocks int
		expected  string
	}{
		{
			name:      "normal case 50%",
			value:     50,
			fgBlock:   "█",
			maxValue:  100,
			bgBlock:   "░",
			maxBlocks: 10,
			expected:  "█████░░░░░",
		},
		{
			name:      "zero value",
			value:     0,
			fgBlock:   "█",
			maxValue:  100,
			bgBlock:   "░",
			maxBlocks: 10,
			expected:  "░░░░░░░░░░",
		},
		{
			name:      "small positive value shows one block",
			value:     0.1,
			fgBlock:   "█",
			maxValue:  100,
			bgBlock:   "░",
			maxBlocks: 10,
			expected:  "█░░░░░░░░░",
		},
		{
			name:      "full value",
			value:     100,
			fgBlock:   "█",
			maxValue:  100,
			bgBlock:   "░",
			maxBlocks: 10,
			expected:  "██████████",
		},
		{
			name:      "value exceeds maxValue - should not panic",
			value:     150,
			fgBlock:   "█",
			maxValue:  100,
			bgBlock:   "░",
			maxBlocks: 10,
			expected:  "██████████",
		},
		{
			name:      "very high value - should not panic",
			value:     1000,
			fgBlock:   "█",
			maxValue:  100,
			bgBlock:   "░",
			maxBlocks: 20,
			expected:  "████████████████████",
		},
		{
			name:      "missing blocks due to float truncation",
			value:     27.5,
			fgBlock:   "▒",
			maxValue:  50,
			bgBlock:   "░",
			maxBlocks: 50,
			expected:  "▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒░░░░░░░░░░░░░░░░░░░░░░░", // Should be 27 fg + 23 bg = 50 total
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HorizontalBar(tt.value, tt.fgBlock, tt.maxValue, tt.bgBlock, tt.maxBlocks)

			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}
