package actuator

import (
	"context"
	"fmt"
	"os/exec"
)

type DirectShutdown struct {
	Mode string
}

func NewDirectShutdown(mode string) *DirectShutdown {
	// Default to sleep if empty
	if mode == "" {
		mode = "off"
	}
	return &DirectShutdown{Mode: mode}
}

func (a *DirectShutdown) Name() string {
	return "DirectShutdown(" + a.Mode + ")"
}

func (a *DirectShutdown) Trigger(ctx context.Context) error {
	var cmd *exec.Cmd

	if a.Mode == "off" {
		// Full Power Off
		cmd = exec.CommandContext(ctx, "shutdown", "-h", "now")
	} else {
		// Default: Suspend to RAM (S3) for WOL compatibility
		// This requires root/privileged access
		cmd = exec.CommandContext(ctx, "sh", "-c", "echo mem > /sys/power/state")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s failed: %v, output: %s", a.Mode, err, string(output))
	}

	return nil
}
