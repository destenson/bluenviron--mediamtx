package recorder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTimestampToDuration(t *testing.T) {
	testCases := []struct {
		name      string
		timestamp int64
		timeScale int
		expected  time.Duration
	}{
		{
			name:      "90kHz_timescale",
			timestamp: 90000,
			timeScale: 90000,
			expected:  time.Second,
		},
		{
			name:      "half_second_90kHz",
			timestamp: 45000,
			timeScale: 90000,
			expected:  500 * time.Millisecond,
		},
		{
			name:      "48kHz_timescale",
			timestamp: 48000,
			timeScale: 48000,
			expected:  time.Second,
		},
		{
			name:      "44.1kHz_timescale",
			timestamp: 44100,
			timeScale: 44100,
			expected:  time.Second,
		},
		{
			name:      "zero",
			timestamp: 0,
			timeScale: 90000,
			expected:  0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := timestampToDuration(tc.timestamp, tc.timeScale)
			require.Equal(t, tc.expected, result)
		})
	}
}
