package tui

import (
	"testing"
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
			expected:    "▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒░░░",
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
