package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/JayYoon0412/pomo/internal/session"
)

var (
	focusMins  int
	breakMins  int
	blockSites []string
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a Pomodoro focus session",
	RunE:  runStart,
}

func init() {
	startCmd.Flags().IntVar(&focusMins, "focus", 25, "Focus duration in minutes")
	startCmd.Flags().IntVar(&breakMins, "break", 5, "Break duration in minutes")
	startCmd.Flags().StringSliceVar(&blockSites, "block", nil, "Comma-separated websites to block during focus (e.g. youtube.com,twitter.com)")
}

func runStart(cmd *cobra.Command, args []string) error {
	if focusMins <= 0 {
		return fmt.Errorf("--focus must be greater than 0, got %d", focusMins)
	}
	if breakMins <= 0 {
		return fmt.Errorf("--break must be greater than 0, got %d", breakMins)
	}

	cfg := session.Config{
		FocusMins:  focusMins,
		BreakMins:  breakMins,
		BlockSites: blockSites,
	}

	if err := session.Run(cfg); err != nil {
		if errors.Is(err, os.ErrPermission) {
			fmt.Fprintln(os.Stderr, "error: permission denied writing to /etc/hosts")
			fmt.Fprintln(os.Stderr, "hint:  run with sudo to enable website blocking")
			os.Exit(1)
		}
		return err
	}
	return nil
}
