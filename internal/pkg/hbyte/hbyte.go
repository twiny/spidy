package hbyte

import (
	"fmt"
	"strings"
)

const (
	b  = "b"
	kb = "kb"
	mb = "mb"
	gb = "gb"
	tb = "tb"
)

// type BYTE int64

const (
	B int64 = 1 << (10 * iota)
	KB
	MB
	GB
	TB
)

// Parse
func Parse(s string) int64 {
	// lower case
	s = strings.ToLower(s)

	var n int64
	var unit string

	fmt.Sscanf(s, "%d%s", &n, &unit)

	switch unit {
	case b:
		return n
	case kb:
		return n * KB
	case mb:
		return n * MB
	case gb:
		return n * GB
	case tb:
		return n * TB
	default:
		return n
	}
}

// String
func String(n int64) string {
	switch {
	case n >= TB:
		return fmt.Sprintf("%d %s", n/TB, tb)
	case n >= GB:
		return fmt.Sprintf("%d %s", n/GB, gb)
	case n >= MB:
		return fmt.Sprintf("%d %s", n/MB, mb)
	case n >= KB:
		return fmt.Sprintf("%d %s", n/KB, kb)
	default:
		return fmt.Sprintf("%d %s", n, b)
	}
}
