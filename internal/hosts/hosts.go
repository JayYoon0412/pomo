package hosts

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	hostsFile  = "/etc/hosts"
	blockStart = "# pomo-block-start"
	blockEnd   = "# pomo-block-end"
	redirectIP = "127.0.0.1"
)

// Block adds pomo-managed entries for each site to the hosts file.
// It first removes any leftover pomo block from a previous interrupted run.
func Block(sites []string) error {
	content, err := os.ReadFile(hostsFile)
	if err != nil {
		return fmt.Errorf("reading %s: %w", hostsFile, err)
	}

	cleaned := removeBlock(string(content))

	var block strings.Builder
	block.WriteString("\n" + blockStart + "\n")
	for _, site := range sites {
		site = strings.ToLower(strings.TrimSpace(site))
		if site == "" {
			continue
		}
		block.WriteString(fmt.Sprintf("%s %s\n", redirectIP, site))
		// Also block the www. variant unless it was already specified
		if !strings.HasPrefix(site, "www.") {
			block.WriteString(fmt.Sprintf("%s www.%s\n", redirectIP, site))
		}
	}
	block.WriteString(blockEnd + "\n")

	if err := os.WriteFile(hostsFile, []byte(cleaned+block.String()), 0644); err != nil {
		return err
	}

	flushDNS()
	return nil
}

// Unblock removes all pomo-managed entries from the hosts file.
func Unblock() error {
	content, err := os.ReadFile(hostsFile)
	if err != nil {
		return fmt.Errorf("reading %s: %w", hostsFile, err)
	}

	cleaned := removeBlock(string(content))

	if err := os.WriteFile(hostsFile, []byte(cleaned), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", hostsFile, err)
	}

	flushDNS()
	return nil
}

// removeBlock strips the pomo-managed block from hosts content.
func removeBlock(content string) string {
	lines := strings.Split(content, "\n")
	var out []string
	inBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == blockStart {
			inBlock = true
			continue
		}
		if trimmed == blockEnd {
			inBlock = false
			continue
		}
		if !inBlock {
			out = append(out, line)
		}
	}

	// Strip trailing blank lines, preserve one trailing newline
	for len(out) > 0 && strings.TrimSpace(out[len(out)-1]) == "" {
		out = out[:len(out)-1]
	}

	return strings.Join(out, "\n") + "\n"
}

// flushDNS attempts to flush the system DNS cache on macOS (best-effort).
func flushDNS() {
	_ = exec.Command("dscacheutil", "-flushcache").Run()
	_ = exec.Command("killall", "-HUP", "mDNSResponder").Run()
}
