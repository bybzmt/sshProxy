package ui

import (
	"fmt"
	"time"
)

const (
	s_kb = 1024
	s_mb = 1024 * 1024
	s_gb = 1024 * 1024 * 1024
)

func FmtSize(t time.Duration, i int64) string {
	num := float64(i) / (float64(t) / float64(time.Second))

	if num > s_gb {
		return fmt.Sprintf("%.2fGB/s", num/s_gb)
	} else if num > s_mb {
		return fmt.Sprintf("%.2fMB/s", num/s_mb)
	} else if num > s_kb {
		return fmt.Sprintf("%.2fKB/s", num/s_kb)
	} else {
		return fmt.Sprintf("%.2fB/s", num)
	}
}
