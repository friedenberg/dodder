package ui

import "fmt"

func GetHumanBytesStringOrError(bytes int64) string {
	if bytes < 0 {
		return fmt.Sprintf("%d bytes (error)", bytes)
	} else {
		return GetHumanBytesString(uint64(bytes))
	}
}

func GetHumanBytesString(bytes uint64) string {
	const unit = uint64(1024)

	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := uint64(unit), 0

	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "kMGTPE"[exp])
}
