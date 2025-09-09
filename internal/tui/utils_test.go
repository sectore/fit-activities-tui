package tui

import (
	"testing"
)

func TestHorizontalStackedBar(t *testing.T) {
	tests := []struct {
		name        string
		value1      uint32
		value1Block string
		value2      uint32
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
