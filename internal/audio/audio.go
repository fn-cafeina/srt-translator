package audio

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func HasFFmpeg() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

func srtTimeToDuration(ts string) (time.Duration, error) {
	ts = strings.Replace(ts, ",", ".", 1)
	parts := strings.Split(ts, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid srt duration format: %s", ts)
	}

	h, _ := strconv.ParseFloat(parts[0], 64)
	m, _ := strconv.ParseFloat(parts[1], 64)
	s, _ := strconv.ParseFloat(parts[2], 64)

	totalSeconds := (h * 3600) + (m * 60) + s
	return time.Duration(totalSeconds * float64(time.Second)), nil
}

func formatDurationFFmpeg(d time.Duration) string {
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	d -= s * time.Second
	ms := d / time.Millisecond

	return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
}

func SliceChunk(inputVideo, startTs, endTs string) (string, error) {
	startDur, err := srtTimeToDuration(startTs)
	if err != nil {
		return "", err
	}

	preBufferedStart := startDur - (10 * time.Second)
	if preBufferedStart < 0 {
		preBufferedStart = 0
	}

	startFmt := formatDurationFFmpeg(preBufferedStart)

	endDur, err := srtTimeToDuration(endTs)
	if err != nil {
		return "", err
	}
	endFmt := formatDurationFFmpeg(endDur)

	return sliceChunkDirect(inputVideo, startFmt, endFmt)
}

func sliceChunkDirect(inputVideo, startTs, endTs string) (string, error) {
	outputFile := filepath.Join(os.TempDir(), fmt.Sprintf("chunk_%d.mp3", time.Now().UnixNano()))

	// Suppress ffmpeg verbosity completely to adhere strictly to UNIX conventions (silent == success)
	cmd := exec.Command("ffmpeg", "-ss", startTs, "-to", endTs, "-i", inputVideo, "-vn", "-ac", "1", "-ar", "16000", "-ab", "32k", "-y", "-loglevel", "quiet", outputFile)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed allocating internal contextual micro-audio (ensure input file has a readable audio track): %w", err)
	}

	return outputFile, nil
}
