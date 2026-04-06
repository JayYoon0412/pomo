package session

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/JayYoon0412/pomo/internal/audio"
	"github.com/JayYoon0412/pomo/internal/hosts"
	"github.com/JayYoon0412/pomo/internal/ui"
)

const sessionsPerCycle = 4

type Config struct {
	FocusMins  int
	BreakMins  int
	BlockSites []string
	SoundPath  string
	SoundName  string
}

func Run(cfg Config) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	disp := ui.NewDisplay()
	sessionNum := 0

	for {
		sessionNum++

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

		if len(cfg.BlockSites) > 0 {
			if err := hosts.Block(cfg.BlockSites); err != nil {
				return err
			}
			sitesBlocked = true
		}

		if cfg.SoundPath != "" {
			player = audio.NewPlayer()
			if err := player.PlayLoop(cfg.SoundPath); err != nil {
				return err
			}
		}

		focusDur := time.Duration(cfg.FocusMins) * time.Minute

		if interrupted := runPhase(disp, ui.PhaseFocus, focusDur, cfg.BlockSites, cfg.SoundName, sessionNum, sigCh, cleanup); interrupted {
			return nil
		}

		// Stop sound and unblock sites before break begins
		cleanup()

		disp.PrintMessage("  Focus complete — starting break...")

		breakDur := time.Duration(cfg.BreakMins) * time.Minute

		if interrupted := runPhase(disp, ui.PhaseBreak, breakDur, nil, "", sessionNum, sigCh, func() {}); interrupted {
			return nil
		}

		disp.PrintMessage(fmt.Sprintf("  Session %d complete!", sessionNum))

		if sessionNum >= sessionsPerCycle {
			fmt.Print("\n  4 sessions complete! Continue? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			resp, _ := reader.ReadString('\n')
			resp = strings.TrimSpace(strings.ToLower(resp))
			if resp != "y" {
				disp.PrintMessage("  Great work! See you next time.")
				return nil
			}
			sessionNum = 0
		}
	}
}

func runPhase(
	disp *ui.Display,
	phase ui.Phase,
	total time.Duration,
	blocked []string,
	sound string,
	sessionNum int,
	sigCh chan os.Signal,
	cleanup func(),
) bool {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	start := time.Now()
	disp.Render(phase, total, total, blocked, sound, sessionNum)

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
				disp.Render(phase, 0, total, blocked, sound, sessionNum)
				return false
			}
			disp.Render(phase, remaining, total, blocked, sound, sessionNum)
		}
	}
}
