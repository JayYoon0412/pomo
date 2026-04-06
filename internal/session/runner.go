package session

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JayYoon0412/pomo/internal/audio"
	"github.com/JayYoon0412/pomo/internal/hosts"
	"github.com/JayYoon0412/pomo/internal/ui"
)

// Config holds the parameters for a Pomodoro session.
type Config struct {
	FocusMins  int
	BreakMins  int
	BlockSites []string
	SoundPath  string
	SoundName  string
}

// Run executes a full Pomodoro session: focus phase followed by break phase.
func Run(cfg Config) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sitesBlocked := false
	var player *audio.Player

	cleanup := func() {
		if player != nil {
			player.Stop()
			player = nil
		}
		if sitesBlocked {
			if err := hosts.Unblock(); err != nil {
				fmt.Fprintf(os.Stderr, "\nwarning: failed to restore /etc/hosts: %v\n", err)
			}
			sitesBlocked = false
		}
	}

	// Block sites for the focus phase
	if len(cfg.BlockSites) > 0 {
		if err := hosts.Block(cfg.BlockSites); err != nil {
			return err
		}
		sitesBlocked = true
	}

	// Start ambient sound for the focus phase
	if cfg.SoundPath != "" {
		player = audio.NewPlayer()
		if err := player.PlayLoop(cfg.SoundPath); err != nil {
			return err
		}
	}

	disp := ui.NewDisplay()
	focusDur := time.Duration(cfg.FocusMins) * time.Minute

	if interrupted := runPhase(disp, ui.PhaseFocus, focusDur, cfg.BlockSites, cfg.SoundName, sigCh, cleanup); interrupted {
		return nil
	}

	// Stop sound and unblock sites before break begins
	cleanup()

	disp.PrintMessage("  Focus complete — starting break...")

	breakDur := time.Duration(cfg.BreakMins) * time.Minute

	if interrupted := runPhase(disp, ui.PhaseBreak, breakDur, nil, "", sigCh, func() {}); interrupted {
		return nil
	}

	disp.PrintMessage("  Break complete. Good work!")
	return nil
}

// runPhase runs a single countdown phase (focus or break).
// Returns true if the phase was cut short by a signal.
func runPhase(
	disp *ui.Display,
	phase ui.Phase,
	total time.Duration,
	blocked []string,
	sound string,
	sigCh chan os.Signal,
	cleanup func(),
) bool {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	start := time.Now()
	disp.Render(phase, total, total, blocked, sound)

	for {
		select {
		case <-sigCh:
			disp.PrintMessage("\n  Interrupted — restoring system state...")
			cleanup()
			os.Exit(0)

		case <-ticker.C:
			elapsed := time.Since(start)
			remaining := total - elapsed
			if remaining <= 0 {
				disp.Render(phase, 0, total, blocked, sound)
				return false
			}
			disp.Render(phase, remaining, total, blocked, sound)
		}
	}
}
