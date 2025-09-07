/*
Copyright © 2025 Ben Ricker

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// parseDurationArg parses a string like "7d", "2m", "1y" and returns a cutoff time.Time from now.
func parseDurationArg(arg string) (time.Time, error) {
	if len(arg) < 2 {
		return time.Time{}, fmt.Errorf("invalid duration argument: %s", arg)
	}
	unit := arg[len(arg)-1]
	numStr := arg[:len(arg)-1]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid number in duration: %s", arg)
	}
	now := time.Now()
	switch unit {
	case 'd':
		return now.AddDate(0, 0, -num), nil
	case 'm':
		return now.AddDate(0, -num, 0), nil
	case 'y':
		return now.AddDate(-num, 0, 0), nil
	default:
		return time.Time{}, fmt.Errorf("invalid unit in duration: %c", unit)
	}
}

// isEmptyLine checks if a string contains only whitespace.
func isEmptyLine(s string) bool {
	return strings.TrimSpace(s) == ""
}

// countLines counts the number of lines in a string, handling edge cases.
func countLines(content string) int {
	lines := strings.Count(content, "\n")
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		lines++
	}
	return lines
}

// previewContent truncates a string to a maximum length for debug logging.
func previewContent(s string) string {
	const contentPreviewLength = 40
	const previewSuffix = "..."
	
	if len(s) > contentPreviewLength {
		return s[:contentPreviewLength] + previewSuffix
	}
	return s
}
