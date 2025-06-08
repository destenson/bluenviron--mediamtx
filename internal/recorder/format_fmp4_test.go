package recorder

import (
	"testing"
	"time"

	"github.com/bluenviron/mediacommon/v2/pkg/formats/fmp4"
	"github.com/stretchr/testify/require"
)

func TestFormatFMP4_FrameContinuity(t *testing.T) {
	// Create a mock recorder instance with logger
	ri := &recorderInstance{
		parent: &mockLogger{},
	}

	// Create formatFMP4 instance
	f := &formatFMP4{
		ri:               ri,
		tracks:           make(map[*formatFMP4Track]*fmp4.InitTrack),
		lastFrameDTS:     make(map[*formatFMP4Track]time.Duration),
		lastFrameTime:    make(map[*formatFMP4Track]time.Time),
		expectedFrameDTS: make(map[*formatFMP4Track]time.Duration),
	}

	// Setup mock track
	track := &formatFMP4Track{
		media: fmp4.MediaTypeVideo,
		initTrack: &fmp4.InitTrack{
			ID:        1,
			TimeScale: 90000,
		},
	}
	f.tracks[track] = track.initTrack

	// Setup mock segment and part
	f.segment = &mockFormatFMP4Segment{
		startDTS: 0,
		startNTP: time.Now(),
		test:     t,
	}

	f.part = &formatFMP4Part{
		s:              f.segment,
		sequenceNumber: 1,
		startDTS:       0,
	}
	f.part.initialize()

	// Test continuous frames
	sampleDuration := uint32(3000) // 30fps at 90000 timeScale
	baseTime := time.Now()

	// Add 5 continuous frames
	for i := 0; i < 5; i++ {
		s := &sample{
			Sample: &fmp4.Sample{
				Duration: sampleDuration,
				Payload:  []byte{0x01, 0x02, 0x03},
			},
			dts: int64(i * int(sampleDuration)),
			ntp: baseTime.Add(time.Duration(i) * 33333333 * time.Nanosecond),
		}

		dts := time.Duration(i * int(sampleDuration) * int(time.Second) / 90000)
		err := f.handleFrame(track, s, dts)
		require.NoError(t, err)
	}

	// Last DTS should be for frame 4
	expectedLastDTS := time.Duration(4 * int(sampleDuration) * int(time.Second) / 90000)
	require.Equal(t, expectedLastDTS, f.lastFrameDTS[track])

	// Now add a frame with a gap - this should be detected
	s := &sample{
		Sample: &fmp4.Sample{
			Duration: sampleDuration,
			Payload:  []byte{0x01, 0x02, 0x03},
		},
		dts: int64(10 * int(sampleDuration)), // Jump to frame 10
		ntp: baseTime.Add(10 * 33333333 * time.Nanosecond),
	}

	// The DTS has a discontinuity - frame 4 to frame 10 (skipping 5 frames)
	discontinuityDTS := time.Duration(10 * int(sampleDuration) * int(time.Second) / 90000)
	err := f.handleFrame(track, s, discontinuityDTS)
	require.NoError(t, err)

	// Last DTS should now be updated to the new value
	require.Equal(t, discontinuityDTS, f.lastFrameDTS[track])
}

func TestFormatFMP4_MultiTrackContinuity(t *testing.T) {
	// Create a mock recorder instance with logger
	ri := &recorderInstance{
		parent: &mockLogger{},
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

	// Setup mock segment and part
	f.segment = &mockFormatFMP4Segment{
		startDTS: 0,
		startNTP: time.Now(),
		test:     t,
	}

	f.part = &formatFMP4Part{
		s:              f.segment,
		sequenceNumber: 1,
		startDTS:       0,
	}
	f.part.initialize()

	// Test with both video and audio frames
	baseTime := time.Now()

	// Add some continuous video frames
	videoDuration := uint32(3000) // 30fps
	for i := 0; i < 3; i++ {
		s := &sample{
			Sample: &fmp4.Sample{
				Duration: videoDuration,
				Payload:  []byte{0x01, 0x02, 0x03},
			},
			dts: int64(i * int(videoDuration)),
			ntp: baseTime.Add(time.Duration(i) * 33333333 * time.Nanosecond),
		}

		dts := time.Duration(i * int(videoDuration) * int(time.Second) / 90000)
		err := f.handleFrame(videoTrack, s, dts)
		require.NoError(t, err)
	}

	// Add some continuous audio frames
	audioDuration := uint32(1000) // 48kHz
	for i := 0; i < 3; i++ {
		s := &sample{
			Sample: &fmp4.Sample{
				Duration: audioDuration,
				Payload:  []byte{0x04, 0x05, 0x06},
			},
			dts: int64(i * int(audioDuration)),
			ntp: baseTime.Add(time.Duration(i) * 20833333 * time.Nanosecond),
		}

		dts := time.Duration(i * int(audioDuration) * int(time.Second) / 48000)
		err := f.handleFrame(audioTrack, s, dts)
		require.NoError(t, err)
	}

	// Verify last DTS values for both tracks
	expectedVideoLastDTS := time.Duration(2 * int(videoDuration) * int(time.Second) / 90000)
	expectedAudioLastDTS := time.Duration(2 * int(audioDuration) * int(time.Second) / 48000)
	
	require.Equal(t, expectedVideoLastDTS, f.lastFrameDTS[videoTrack])
	require.Equal(t, expectedAudioLastDTS, f.lastFrameDTS[audioTrack])

	// Now add discontinuous frames for both tracks
	// Video discontinuity
	sVideo := &sample{
		Sample: &fmp4.Sample{
			Duration: videoDuration,
			Payload:  []byte{0x01, 0x02, 0x03},
		},
		dts: int64(10 * int(videoDuration)), // Jump to frame 10
		ntp: baseTime.Add(10 * 33333333 * time.Nanosecond),
	}
	videoDiscontinuityDTS := time.Duration(10 * int(videoDuration) * int(time.Second) / 90000)
	err := f.handleFrame(videoTrack, sVideo, videoDiscontinuityDTS)
	require.NoError(t, err)

	// Audio discontinuity
	sAudio := &sample{
		Sample: &fmp4.Sample{
			Duration: audioDuration,
			Payload:  []byte{0x04, 0x05, 0x06},
		},
		dts: int64(10 * int(audioDuration)), // Jump to frame 10
		ntp: baseTime.Add(10 * 20833333 * time.Nanosecond),
	}
	audioDiscontinuityDTS := time.Duration(10 * int(audioDuration) * int(time.Second) / 48000)
	err = f.handleFrame(audioTrack, sAudio, audioDiscontinuityDTS)
	require.NoError(t, err)

	// Verify updated last DTS values
	require.Equal(t, videoDiscontinuityDTS, f.lastFrameDTS[videoTrack])
	require.Equal(t, audioDiscontinuityDTS, f.lastFrameDTS[audioTrack])
}
