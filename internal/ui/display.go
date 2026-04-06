package ui

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const barWidth = 32

// Phase represents the current timer phase.
type Phase int

const (
	PhaseFocus Phase = iota
	PhaseBreak
)

// ANSI escape codes
const (
	ansiReset   = "\033[0m"
	ansiBold    = "\033[1m"
	ansiDim     = "\033[2m"
	ansiCyan    = "\033[36m"
	ansiGreen   = "\033[32m"
	ansiClearLn = "\033[2K"
)

// Display manages terminal rendering for the session timer.
type Display struct {
	prevLines int
}

func NewDisplay() *Display {
	return &Display{}
}

// Render draws (or redraws) the timer display in place.
func (d *Display) Render(phase Phase, remaining, total time.Duration, blocked []string, sound string, sessionNum int) {
	lines := d.buildLines(phase, remaining, total, blocked, sound, sessionNum)

	// Move cursor back to the top of the previous render
	if d.prevLines > 0 {
		fmt.Fprintf(os.Stdout, "\033[%dA", d.prevLines)
	}

	for _, line := range lines {
		fmt.Fprintf(os.Stdout, "\r%s%s\n", ansiClearLn, line)
	}

	d.prevLines = len(lines)
}

// PrintMessage prints a status message below the current render area
// and resets the line counter so the next Render starts fresh.
func (d *Display) PrintMessage(msg string) {
	d.prevLines = 0
	fmt.Println(msg)
}

func (d *Display) buildLines(phase Phase, remaining, total time.Duration, blocked []string, sound string, sessionNum int) []string {
	var lines []string

	lines = append(lines, "") // top padding

	// Session indicator
	lines = append(lines, fmt.Sprintf("  %sSession %d / 4%s", ansiDim, sessionNum, ansiReset))

	// Phase header
	if phase == PhaseFocus {
		lines = append(lines, fmt.Sprintf("  %s%s● 🍅 FOCUS%s", ansiBold, ansiCyan, ansiReset))
	} else {
		lines = append(lines, fmt.Sprintf("  %s%s○ ☕️ BREAK%s", ansiBold, ansiGreen, ansiReset))
	}

	// Countdown
	remaining = remaining.Round(time.Second)
	if remaining < 0 {
		remaining = 0
	}
	lines = append(lines, fmt.Sprintf("  %s%s%s remaining", ansiBold, formatDuration(remaining), ansiReset))

	// Progress bar
	var progress float64
	if total > 0 {
		progress = float64(total-remaining) / float64(total)
	}
	lines = append(lines, "  "+renderBar(progress, barWidth))

	// Blocked sites (only shown during focus)
	if len(blocked) > 0 && phase == PhaseFocus {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  %sblocking:%s %s", ansiDim, ansiReset, strings.Join(blocked, "  ")))
	}

	// Ambient sound (only shown during focus)
	if sound != "" && phase == PhaseFocus {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  %s🎵 %s%s", ansiDim, sound, ansiReset))
	}

	lines = append(lines, "") // bottom padding

	return lines
}

func formatDuration(d time.Duration) string {
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

func renderBar(progress float64, width int) string {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	filled := int(progress * float64(width))
	empty := width - filled
	return ansiCyan + strings.Repeat("█", filled) + ansiDim + strings.Repeat("░", empty) + ansiReset
}
