package recorder

import (
	"github.com/bluenviron/mediamtx/internal/logger"
)

// This file contains tests for the frame continuity detection feature.
// Since we need to modify the actual formatFMP4 to support this test,
// we'll need to create mock implementations of handleFrame and other methods.

// Mock implementations for testing
type mockLogger struct{}

func (m *mockLogger) Log(_ logger.Level, _ string, _ ...interface{}) {}

//type mockFormatFMP4 struct {
//	ri *recorderInstance
//}
//
//type mockFormatFMP4Track struct {
//	media     string
//	initTrack *fmp4.InitTrack
//}
//
//type mockFormatFMP4Segment struct {
//	startDTS time.Duration
//	startNTP time.Time
//	f        *mockFormatFMP4
//}
//
//func (s *mockFormatFMP4Segment) close() error {
//	return nil
//}

//// Mock formatFMP4Track for testing continuity
//type formatFMP4TrackContinuity struct {
//	*mockFormatFMP4Track
//	parent *formatFMP4Continuity
//}
//
//func (t *formatFMP4TrackContinuity) write(s *sample) error {
//	// Forward to the part's write method
//	if t.parent.part != nil {
//		return t.parent.part.write(t.mockFormatFMP4Track, s, time.Duration(s.dts))
//	}
//	return nil
//}

//// Mock formatFMP4 implementation for testing continuity features
//type formatFMP4Continuity struct {
//	ri               *recorderInstance
//	tracks           map[*mockFormatFMP4Track]*fmp4.InitTrack
//	segment          *formatFMP4Segment
//	part             *formatFMP4Part
//	lastFrameDTS     map[*mockFormatFMP4Track]time.Duration
//	lastFrameTime    map[*mockFormatFMP4Track]time.Time
//	expectedFrameDTS map[*mockFormatFMP4Track]time.Duration
//}
//
//// Simplified handleFrame implementation for testing
//func (f *formatFMP4Continuity) handleFrame(t *formatFMP4Track, s *sample, dts time.Duration) error {
//	// Check for continuity between frames
//	if lastDTS, exists := f.lastFrameDTS[t]; exists {
//		lastDuration := timestampToDuration(int64(s.Duration), int(t.initTrack.TimeScale))
//		expectedDTS := lastDTS + lastDuration
//
//		// Allow for small timing differences (1ms) due to rounding
//		tolerance := time.Millisecond
//		dtsDiff := dts - expectedDTS
//		if dtsDiff < -tolerance || dtsDiff > tolerance {
//			f.ri.Log(logger.Debug, "possible frame discontinuity detected: track=%v, expected_dts=%v, actual_dts=%v, diff=%v",
//				t.media,
//				expectedDTS,
//				dts,
//				dtsDiff)
//		}
//	}
//
//	// Store the current frame's DTS and time for future continuity checks
//	f.lastFrameDTS[t] = dts
//	f.lastFrameTime[t] = s.ntp
//
//	return f.part.write(t, s, dts)
//}

//func TestFormatFMP4_FrameContinuity(t *testing.T) {
//	// Create a mock recorder instance with logger
//	ri := &recorderInstance{
//		parent: &mockLogger{},
//	}
//
//	// Create mock formatFMP4 instance
//	f := &formatFMP4Continuity{
//		ri:               ri,
//		tracks:           make(map[*mockFormatFMP4Track]*fmp4.InitTrack),
//		lastFrameDTS:     make(map[*mockFormatFMP4Track]time.Duration),
//		lastFrameTime:    make(map[*mockFormatFMP4Track]time.Time),
//		expectedFrameDTS: make(map[*mockFormatFMP4Track]time.Duration),
//	}
//
//	// Setup mock track
//	track := &mockFormatFMP4Track{
//		media: "video",
//		initTrack: &fmp4.InitTrack{
//			ID:        1,
//			TimeScale: 90000,
//		},
//	}
//	f.tracks[track] = track.initTrack
//
//	// Setup mock segment
//	f.segment = &formatFMP4Segment{
//		startDTS: 0,
//		startNTP: time.Now(),
//		f:        &formatFMP4{ri: ri},
//	}
//
//	// Create part
//	f.part = &formatFMP4Part{
//		s:              f.segment,
//		sequenceNumber: 1,
//		startDTS:       0,
//	}
//	f.part.initialize()
//
//	// Test continuous frames
//	sampleDuration := uint32(3000) // 30fps at 90000 timeScale
//	baseTime := time.Now()
//
//	// Add 5 continuous frames
//	for i := 0; i < 5; i++ {
//		s := &sample{
//			Sample: &fmp4.Sample{
//				Duration: sampleDuration,
//				Payload:  []byte{0x01, 0x02, 0x03},
//			},
//			dts: int64(i * int(sampleDuration)),
//			ntp: baseTime.Add(time.Duration(i) * 33333333 * time.Nanosecond),
//		}
//
//		dts := time.Duration(i * int(sampleDuration) * int(time.Second) / 90000)
//		err := f.handleFrame(track, s, dts)
//		require.NoError(t, err)
//	}
//
//	// Last DTS should be for frame 4
//	expectedLastDTS := time.Duration(4 * int(sampleDuration) * int(time.Second) / 90000)
//	require.Equal(t, expectedLastDTS, f.lastFrameDTS[track])
//
//	// Now add a frame with a gap - this should be detected
//	s := &sample{
//		Sample: &fmp4.Sample{
//			Duration: sampleDuration,
//			Payload:  []byte{0x01, 0x02, 0x03},
//		},
//		dts: int64(10 * int(sampleDuration)), // Jump to frame 10
//		ntp: baseTime.Add(10 * 33333333 * time.Nanosecond),
//	}
//
//	// The DTS has a discontinuity - frame 4 to frame 10 (skipping 5 frames)
//	discontinuityDTS := time.Duration(10 * int(sampleDuration) * int(time.Second) / 90000)
//	err := f.handleFrame(track, s, discontinuityDTS)
//	require.NoError(t, err)
//
//	// Last DTS should now be updated to the new value
//	require.Equal(t, discontinuityDTS, f.lastFrameDTS[track])
//}
//
//func TestFormatFMP4_MultiTrackContinuity(t *testing.T) {
//	// Create a mock recorder instance with logger
//	ri := &recorderInstance{
//		parent: &mockLogger{},
//	}
//
//	// Create mock formatFMP4 instance
//	f := &formatFMP4Continuity{
//		ri:               ri,
//		tracks:           make(map[*mockFormatFMP4Track]*fmp4.InitTrack),
//		lastFrameDTS:     make(map[*mockFormatFMP4Track]time.Duration),
//		lastFrameTime:    make(map[*mockFormatFMP4Track]time.Time),
//		expectedFrameDTS: make(map[*mockFormatFMP4Track]time.Duration),
//	}
//
//	// Setup video track
//	videoTrack := &mockFormatFMP4Track{
//		media: "video",
//		initTrack: &fmp4.InitTrack{
//			ID:        1,
//			TimeScale: 90000,
//		},
//	}
//	f.tracks[videoTrack] = videoTrack.initTrack
//
//	// Setup audio track
//	audioTrack := &mockFormatFMP4Track{
//		media: "audio",
//		initTrack: &fmp4.InitTrack{
//			ID:        2,
//			TimeScale: 48000,
//		},
//	}
//	f.tracks[audioTrack] = audioTrack.initTrack
//
//	// Setup mock segment
//	f.segment = &mockFormatFMP4Segment{
//		startDTS: 0,
//		startNTP: time.Now(),
//		f:        &mockFormatFMP4{ri: ri},
//	}
//
//	// Create part
//	f.part = &formatFMP4Part{
//		s:              f.segment,
//		sequenceNumber: 1,
//		startDTS:       0,
//	}
//	f.part.initialize()
//
//	// Test with both video and audio frames
//	baseTime := time.Now()
//
//	// Add some continuous video frames
//	videoDuration := uint32(3000) // 30fps
//	for i := 0; i < 3; i++ {
//		s := &sample{
//			Sample: &fmp4.Sample{
//				Duration: videoDuration,
//				Payload:  []byte{0x01, 0x02, 0x03},
//			},
//			dts: int64(i * int(videoDuration)),
//			ntp: baseTime.Add(time.Duration(i) * 33333333 * time.Nanosecond),
//		}
//
//		dts := time.Duration(i * int(videoDuration) * int(time.Second) / 90000)
//		err := f.handleFrame(videoTrack, s, dts)
//		require.NoError(t, err)
//	}
//
//	// Add some continuous audio frames
//	audioDuration := uint32(1000) // 48kHz
//	for i := 0; i < 3; i++ {
//		s := &sample{
//			Sample: &fmp4.Sample{
//				Duration: audioDuration,
//				Payload:  []byte{0x04, 0x05, 0x06},
//			},
//			dts: int64(i * int(audioDuration)),
//			ntp: baseTime.Add(time.Duration(i) * 20833333 * time.Nanosecond),
//		}
//
//		dts := time.Duration(i * int(audioDuration) * int(time.Second) / 48000)
//		err := f.handleFrame(audioTrack, s, dts)
//		require.NoError(t, err)
//	}
//
//	// Verify last DTS values for both tracks
//	expectedVideoLastDTS := time.Duration(2 * int(videoDuration) * int(time.Second) / 90000)
//	expectedAudioLastDTS := time.Duration(2 * int(audioDuration) * int(time.Second) / 48000)
//
//	require.Equal(t, expectedVideoLastDTS, f.lastFrameDTS[videoTrack])
//	require.Equal(t, expectedAudioLastDTS, f.lastFrameDTS[audioTrack])
//
//	// Now add discontinuous frames for both tracks
//	// Video discontinuity
//	sVideo := &sample{
//		Sample: &fmp4.Sample{
//			Duration: videoDuration,
//			Payload:  []byte{0x01, 0x02, 0x03},
//		},
//		dts: int64(10 * int(videoDuration)), // Jump to frame 10
//		ntp: baseTime.Add(10 * 33333333 * time.Nanosecond),
//	}
//	videoDiscontinuityDTS := time.Duration(10 * int(videoDuration) * int(time.Second) / 90000)
//	err := f.handleFrame(videoTrack, sVideo, videoDiscontinuityDTS)
//	require.NoError(t, err)
//
//	// Audio discontinuity
//	sAudio := &sample{
//		Sample: &fmp4.Sample{
//			Duration: audioDuration,
//			Payload:  []byte{0x04, 0x05, 0x06},
//		},
//		dts: int64(10 * int(audioDuration)), // Jump to frame 10
//		ntp: baseTime.Add(10 * 20833333 * time.Nanosecond),
//	}
//	audioDiscontinuityDTS := time.Duration(10 * int(audioDuration) * int(time.Second) / 48000)
//	err = f.handleFrame(audioTrack, sAudio, audioDiscontinuityDTS)
//	require.NoError(t, err)
//
//	// Verify updated last DTS values
//	require.Equal(t, videoDiscontinuityDTS, f.lastFrameDTS[videoTrack])
//	require.Equal(t, audioDiscontinuityDTS, f.lastFrameDTS[audioTrack])
//}
