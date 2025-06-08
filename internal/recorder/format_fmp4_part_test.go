package recorder

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bluenviron/mediacommon/v2/pkg/formats/fmp4"
	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediamtx/internal/conf"
	"github.com/bluenviron/mediamtx/internal/logger"
)

type mockFormatFMP4Track struct {
	media     fmp4.MediaType
	initTrack *fmp4.InitTrack
}

type mockLogger struct{}

func (m *mockLogger) Log(_ logger.Level, _ string, _ ...interface{}) {}

// Mock formatFMP4Segment for testing
type mockFormatFMP4Segment struct {
	path      string
	fi        *os.File
	startNTP  time.Time
	startDTS  time.Duration
	partCount uint32

	// Add reference to test instance
	test *testing.T
}

func (s *mockFormatFMP4Segment) nextPartID() uint32 {
	s.partCount++
	return s.partCount
}

func (s *mockFormatFMP4Segment) close() error {
	if s.fi != nil {
		return s.fi.Close()
	}
	return nil
}

func (s *mockFormatFMP4Segment) duration() time.Duration {
	return time.Duration(0)
}

func TestFormatFMP4Part_CalculateExpectedFrameCount(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "fmp4-part-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Setup mock track
	track := &mockFormatFMP4Track{
		media: "video",
		initTrack: &fmp4.InitTrack{
			ID:        1,
			TimeScale: 90000, // Typical value for video
		},
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
			// Create mock segment
			segment := &mockFormatFMP4Segment{
				path:     filepath.Join(tmpDir, "test.mp4"),
				startNTP: time.Now(),
				startDTS: 0,
				test:     t,
			}

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

	// Setup mock track
	track := &mockFormatFMP4Track{
		media: "video",
		initTrack: &fmp4.InitTrack{
			ID:        1,
			TimeScale: 90000, // Typical for video
		},
	}

	// Create mock segment
	segment := &mockFormatFMP4Segment{
		path:     filepath.Join(tmpDir, "test.mp4"),
		startNTP: time.Now(),
		startDTS: 0,
		test:     t,
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

	// Test with various sample durations and timescales
	tests := []struct {
		name          string
		timeScale     uint32
		sampleDuration uint32
		expectedFPS   float64
	}{
		{
			name:          "90kHz_3000_30fps",
			timeScale:     90000,
			sampleDuration: 3000,
			expectedFPS:   30.0,
		},
		{
			name:          "48kHz_1000_48fps_audio",
			timeScale:     48000,
			sampleDuration: 1000,
			expectedFPS:   48.0,
		},
		{
			name:          "90kHz_1500_60fps",
			timeScale:     90000,
			sampleDuration: 1500,
			expectedFPS:   60.0,
		},
		{
			name:          "25kHz_1000_25fps",
			timeScale:     25000,
			sampleDuration: 1000,
			expectedFPS:   25.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock track
			track := &mockFormatFMP4Track{
				media: "video",
				initTrack: &fmp4.InitTrack{
					ID:        1,
					TimeScale: tt.timeScale,
				},
			}

			// Create mock segment
			segment := &mockFormatFMP4Segment{
				path:     filepath.Join(tmpDir, "test.mp4"),
				startNTP: time.Now(),
				startDTS: 0,
				test:     t,
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

func TestFormatFMP4Part_Integration(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "fmp4-part-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Mock recorder instance
	ri := &recorderInstance{
		pathFormat:     "%Y-%m-%d_%H-%M-%S.mp4",
		format:         conf.RecordFormatFMP4,
		partDuration:   2 * time.Second,
		segmentDuration: 10 * time.Second,
		pathName:       "test",
		parent:         &mockLogger{},
		onSegmentCreate: func(string) {},
		onSegmentComplete: func(string, time.Duration) {},
	}

	// Create formatFMP4 instance
	f := &formatFMP4{
		ri:               ri,
		tracks:           make(map[*formatFMP4Track]*fmp4.InitTrack),
		lastFrameDTS:     make(map[*formatFMP4Track]time.Duration),
		lastFrameTime:    make(map[*formatFMP4Track]time.Time),
		expectedFrameDTS: make(map[*formatFMP4Track]time.Duration),
	}

	// Setup video track
	videoTrack := &formatFMP4Track{
		media: fmp4.MediaTypeVideo,
		initTrack: &fmp4.InitTrack{
			ID:        1,
			TimeScale: 90000,
		},
	}
	f.tracks[videoTrack] = videoTrack.initTrack

	// Setup audio track
	audioTrack := &formatFMP4Track{
		media: fmp4.MediaTypeAudio,
		initTrack: &fmp4.InitTrack{
			ID:        2,
			TimeScale: 48000,
		},
	}
	f.tracks[audioTrack] = audioTrack.initTrack

	// Create segment
	f.segment = &formatFMP4Segment{
		f:        f,
		startDTS: 0,
		startNTP: time.Now(),
		path:     filepath.Join(tmpDir, "test.mp4"),
	}

	// Create part
	f.part = &formatFMP4Part{
		s:              f.segment,
		sequenceNumber: 1,
		startDTS:       0,
	}
	f.part.initialize()

	// Generate video frames at 30fps for 3 seconds
	// This should exceed our 2-second part duration
	videoDuration := uint32(3000) // 30fps
	videoFrameCount := 90         // 3 seconds of 30fps
	
	// For the first 60 frames (2 seconds), we should stay in part 1
	// After that, we should create a new part
	for i := 0; i < videoFrameCount; i++ {
		s := &sample{
			Sample: &fmp4.Sample{
				Duration: videoDuration,
				Payload:  []byte{0x01, 0x02, 0x03},
			},
			dts: int64(i * int(videoDuration)),
			ntp: time.Now().Add(time.Duration(i) * 33333333 * time.Nanosecond),
		}

		dts := time.Duration(i * int(videoDuration) * int(time.Second) / 90000)
		
		// Every 30 frames (1 second), add audio samples too
		if i%30 == 0 {
			audioS := &sample{
				Sample: &fmp4.Sample{
					Duration: 1000, // 48kHz audio sample
					Payload:  []byte{0x04, 0x05, 0x06},
				},
				dts: int64(i/30) * 48000,
				ntp: time.Now().Add(time.Duration(i) * 33333333 * time.Nanosecond),
			}
			
			audioDts := time.Duration(i/30) * time.Second
			err := f.handleFrame(audioTrack, audioS, audioDts)
			require.NoError(t, err)
		}
		
		// Process video frame
		err := f.handleFrame(videoTrack, s, dts)
		require.NoError(t, err)
		
		// At frame 60, we should be in the second part
		if i == 60 {
			require.Equal(t, uint32(2), f.part.sequenceNumber, 
				"Expected to be in the second part after 60 frames (2 seconds)")
		}
	}
	
	// Verify frame counts in the final part
	expectedVideoFrames := videoFrameCount - 60 // 30
	require.Equal(t, expectedVideoFrames, f.part.trackFrameCounts[videoTrack],
		"Incorrect video frame count in the second part")
}
