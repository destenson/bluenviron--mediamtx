package recorder

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bluenviron/mediacommon/v2/pkg/formats/fmp4"
	"github.com/stretchr/testify/require"
)

// MediaType constants for testing
const (
	MediaTypeVideo = "video"
	MediaTypeAudio = "audio"
)

func TestFormatFMP4Part_CalculateExpectedFrameCount(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "fmp4-part-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Setup recorder instance
	ri := &recorderInstance{
		parent: &mockLogger{},
	}

	// Create formatFMP4 instance
	f := &formatFMP4{
		ri:     ri,
		tracks: []*formatFMP4Track{},
	}

	// Setup track
	track := &formatFMP4Track{
		f: f,
		initTrack: &fmp4.InitTrack{
			ID:        1,
			TimeScale: 90000, // Typical value for video
		},
	}

	// Setup segment
	segment := &formatFMP4Segment{
		f:        f,
		path:     filepath.Join(tmpDir, "test.mp4"),
		startNTP: time.Now(),
		startDTS: 0,
	}

	tests := []struct {
		name          string
		frameRate     float64
		partDuration  time.Duration
		expectedCount int
	}{
		{
			name:          "30fps_2sec",
			frameRate:     30.0,
			partDuration:  2 * time.Second,
			expectedCount: 60,
		},
		{
			name:          "60fps_1sec",
			frameRate:     60.0,
			partDuration:  1 * time.Second,
			expectedCount: 60,
		},
		{
			name:          "25fps_4sec",
			frameRate:     25.0,
			partDuration:  4 * time.Second,
			expectedCount: 100,
		},
		{
			name:          "zero_duration",
			frameRate:     30.0,
			partDuration:  0,
			expectedCount: 1, // Should default to at least 1 frame
		},
		{
			name:          "zero_framerate",
			frameRate:     0,
			partDuration:  2 * time.Second,
			expectedCount: 60, // Should use default 30fps
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create part
			part := &formatFMP4Part{
				s:              segment,
				sequenceNumber: 1,
				startDTS:       0,
			}
			part.initialize()

			// Set frame rate
			if tt.frameRate > 0 {
				part.frameRates[track] = tt.frameRate
			}

			// Calculate expected frame count
			count := part.calculateExpectedFrameCount(track, tt.partDuration)
			require.Equal(t, tt.expectedCount, count)
		})
	}
}

func TestFormatFMP4Part_FrameCounting(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "fmp4-part-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Setup recorder instance
	ri := &recorderInstance{
		parent: &mockLogger{},
	}

	// Create formatFMP4 instance
	f := &formatFMP4{
		ri:     ri,
		tracks: []*formatFMP4Track{},
	}

	// Setup track
	track := &formatFMP4Track{
		f: f,
		initTrack: &fmp4.InitTrack{
			ID:        1,
			TimeScale: 90000, // Typical for video
		},
	}

	// Setup segment
	segment := &formatFMP4Segment{
		f:        f,
		path:     filepath.Join(tmpDir, "test.mp4"),
		startNTP: time.Now(),
		startDTS: 0,
	}

	// Create part
	part := &formatFMP4Part{
		s:              segment,
		sequenceNumber: 1,
		startDTS:       0,
	}
	part.initialize()

	// Create sample with fixed duration (3000 timeScale units = 30fps at 90000 timeScale)
	sampleDuration := uint32(3000)
	expectedFrameRate := float64(90000) / float64(sampleDuration) // Should be 30fps

	// Add several frames and check frame counting
	frameCount := 10
	for i := 0; i < frameCount; i++ {
		s := &sample{
			Sample: &fmp4.Sample{
				Duration: sampleDuration,
				Payload:  []byte{0x01, 0x02, 0x03}, // Dummy payload
			},
			dts: int64(i * int(sampleDuration)),
			ntp: time.Now(),
		}

		dts := time.Duration(i * int(sampleDuration) * int(time.Second) / 90000)
		err := part.write(track, s, dts)
		require.NoError(t, err)
	}

	// Verify frame count and frame rate calculation
	require.Equal(t, frameCount, part.trackFrameCounts[track])
	require.InDelta(t, expectedFrameRate, part.frameRates[track], 0.001)

	// Verify duration calculation
	expectedDuration := time.Duration(frameCount * int(sampleDuration) * int(time.Second) / 90000)
	require.Equal(t, expectedDuration, part.duration())
}

func TestFormatFMP4Part_FrameRateCalculation(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "fmp4-part-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Setup recorder instance
	ri := &recorderInstance{
		parent: &mockLogger{},
	}

	// Create formatFMP4 instance
	f := &formatFMP4{
		ri:     ri,
		tracks: []*formatFMP4Track{},
	}

	// Test with various sample durations and timescales
	tests := []struct {
		name           string
		timeScale      uint32
		sampleDuration uint32
		expectedFPS    float64
	}{
		{
			name:           "90kHz_3000_30fps",
			timeScale:      90000,
			sampleDuration: 3000,
			expectedFPS:    30.0,
		},
		{
			name:           "48kHz_1000_48fps_audio",
			timeScale:      48000,
			sampleDuration: 1000,
			expectedFPS:    48.0,
		},
		{
			name:           "90kHz_1500_60fps",
			timeScale:      90000,
			sampleDuration: 1500,
			expectedFPS:    60.0,
		},
		{
			name:           "25kHz_1000_25fps",
			timeScale:      25000,
			sampleDuration: 1000,
			expectedFPS:    25.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup track
			track := &formatFMP4Track{
				f: f,
				initTrack: &fmp4.InitTrack{
					ID:        1,
					TimeScale: tt.timeScale,
				},
			}

			// Setup segment
			segment := &formatFMP4Segment{
				f:        f,
				path:     filepath.Join(tmpDir, "test.mp4"),
				startNTP: time.Now(),
				startDTS: 0,
			}

			// Create part
			part := &formatFMP4Part{
				s:              segment,
				sequenceNumber: 1,
				startDTS:       0,
			}
			part.initialize()

			// Add a frame to calculate frame rate
			s := &sample{
				Sample: &fmp4.Sample{
					Duration: tt.sampleDuration,
					Payload:  []byte{0x01, 0x02, 0x03}, // Dummy payload
				},
				dts: 0,
				ntp: time.Now(),
			}

			err := part.write(track, s, 0)
			require.NoError(t, err)

			// Verify frame rate calculation
			require.InDelta(t, tt.expectedFPS, part.frameRates[track], 0.001)
		})
	}
}

func TestFormatFMP4Part_CalculateExpectedFrameCount2(t *testing.T) {
	// Create a formatFMP4Part for testing
	part := &formatFMP4Part{}
	part.initialize()

	// Create a test track
	track := &formatFMP4Track{
		initTrack: &fmp4.InitTrack{
			ID:        1,
			TimeScale: 90000,
		},
	}

	// Test with known frame rate
	part.frameRates[track] = 30.0 // 30fps
	expectedFrames := part.calculateExpectedFrameCount(track, 2*time.Second)
	require.Equal(t, 60, expectedFrames, "Should expect 60 frames for 2 seconds at 30fps")

	// Test with unknown frame rate (should default to 30fps)
	unknownTrack := &formatFMP4Track{
		initTrack: &fmp4.InitTrack{
			ID:        2,
			TimeScale: 90000,
		},
	}
	expectedFrames = part.calculateExpectedFrameCount(unknownTrack, 2*time.Second)
	require.Equal(t, 60, expectedFrames, "Should default to 30fps when frame rate is unknown")

	// Test with very short duration (should always return at least 1 frame)
	expectedFrames = part.calculateExpectedFrameCount(track, 1*time.Millisecond)
	require.Equal(t, 1, expectedFrames, "Should return at least 1 frame even for short durations")
}

func TestFormatFMP4Part_Duration(t *testing.T) {
	// Create a formatFMP4Part for testing
	part := &formatFMP4Part{
		startDTS: 10 * time.Second,
		endDTS:   15 * time.Second,
	}

	// Test duration calculation
	duration := part.duration()
	require.Equal(t, 5*time.Second, duration, "Duration should be endDTS - startDTS")
}

// Skip the integration test as it's too complex for this fix
