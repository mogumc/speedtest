package utils

import "fmt"

func BandwidthToGbps(bandwidth int64) float64 {
	return float64(bandwidth) * 8.0 / 1_000_000_000
}

func ReadableSize(value float64) string {
	units := []string{"KB/s", "MB/s", "GB/s", "TB/s"}
	var i int
	for value >= 1024 && i < len(units)-1 {
		value /= 1024
		i++
	}
	switch {
	case value >= 100:
		return fmt.Sprintf("%.0f %s", value, units[i])
	case value >= 10:
		return fmt.Sprintf("%.1f %s", value, units[i])
	default:
		return fmt.Sprintf("%.2f %s", value, units[i])
	}
}
