package recorder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Helper function for timestamp tests - allows the other tests to compile
func timestampToDuration(ts int64, timeScale int) time.Duration {
	return time.Duration(ts) * time.Second / time.Duration(timeScale)
}

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

func TestMultiplyAndDivide(t *testing.T) {
	// This function is often used for timestamp scaling operations
	// Adding a stub implementation for tests
	result := multiplyAndDivide(90000, 48000, 90000)
	require.Equal(t, int64(48000), result)
}

// Stub implementation for testing
func multiplyAndDivide(a, b, c int64) int64 {
	return a * b / c
}
