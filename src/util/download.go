package util

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// DownloadProgress tracks progress for file downloads
type DownloadProgress struct {
	total      int64
	downloaded int64
	message    string
	width      int
	mu         sync.Mutex
	writer     io.Writer
	startTime  time.Time
}

// NewDownloadProgress creates a new download progress tracker
func NewDownloadProgress(total int64, message string) *DownloadProgress {
	return &DownloadProgress{
		total:     total,
		width:     30,
		message:   message,
		writer:    os.Stderr,
		startTime: time.Now(),
	}
}

// Update updates the download progress with bytes downloaded
func (dp *DownloadProgress) Update(downloaded int64) {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	dp.downloaded = downloaded
	dp.render()
}

// FormatBytes formats bytes as human-readable string
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func (dp *DownloadProgress) render() {
	if dp.total <= 0 {
		// Unknown size - show bytes downloaded only
		fmt.Fprintf(dp.writer, "\r%s %s downloaded", dp.message, FormatBytes(dp.downloaded))
		return
	}

	percent := float64(dp.downloaded) / float64(dp.total)
	filled := int(percent * float64(dp.width))
	if filled > dp.width {
		filled = dp.width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", dp.width-filled)

	// Calculate speed
	elapsed := time.Since(dp.startTime).Seconds()
	var speed string
	if elapsed > 0 {
		bytesPerSec := float64(dp.downloaded) / elapsed
		speed = fmt.Sprintf("%s/s", FormatBytes(int64(bytesPerSec)))
	}

	fmt.Fprintf(dp.writer, "\r%s [%s] %3.0f%% %s/%s %s",
		dp.message, bar, percent*100,
		FormatBytes(dp.downloaded), FormatBytes(dp.total), speed)
}

// Complete marks the download as complete
func (dp *DownloadProgress) Complete() {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	dp.downloaded = dp.total
	dp.render()
	elapsed := time.Since(dp.startTime).Round(time.Millisecond)
	fmt.Fprintf(dp.writer, "\n%s Downloaded %s in %s\n", SuccessIcon, FormatBytes(dp.total), elapsed)
}

// Fail marks the download as failed
func (dp *DownloadProgress) Fail(message string) {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	fmt.Fprintf(dp.writer, "\n%s %s\n", ErrorIcon, message)
}

// ProgressReader wraps an io.Reader to track progress
type ProgressReader struct {
	reader   io.Reader
	progress *DownloadProgress
	read     int64
}

// NewProgressReader creates a reader that tracks download progress
func NewProgressReader(reader io.Reader, total int64, message string) *ProgressReader {
	return &ProgressReader{
		reader:   reader,
		progress: NewDownloadProgress(total, message),
	}
}

// Read implements io.Reader and updates progress
func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.read += int64(n)
		pr.progress.Update(pr.read)
	}
	return n, err
}

// Complete marks the download as complete
func (pr *ProgressReader) Complete() {
	pr.progress.Complete()
}

// Fail marks the download as failed
func (pr *ProgressReader) Fail(message string) {
	pr.progress.Fail(message)
}
