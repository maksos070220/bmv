package cli

import (
	"fmt"
	"io"
	"time"
)

func renderProgress(out io.Writer, start time.Time, downloaded int64, total int64) {
	elapsed := time.Since(start).Seconds()
	if elapsed < 0.001 {
		elapsed = 0.001
	}

	speed := float64(downloaded) / elapsed / 1024 / 1024
	downloadedMB := float64(downloaded) / 1024 / 1024

	if total <= 0 {
		_, _ = fmt.Fprintf(out, "\rDownloaded %.1f MB | %.1f MB/s", downloadedMB, speed)
		return
	}

	percent := float64(downloaded) / float64(total) * 100

	const width = 30
	filled := int(percent / 100 * width)

	if filled > width {
		filled = width
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	remaining := float64(total-downloaded) / 1024 / 1024
	eta := remaining / speed

	_, _ = fmt.Fprintf(
		out,
		"\r[%s] %5.1f%% %.1f/%.1f MB %.1f MB/s ETA %s",
		bar,
		percent,
		downloadedMB,
		float64(total)/1024/1024,
		speed,
		time.Duration(eta*float64(time.Second)).Round(time.Second),
	)
}
