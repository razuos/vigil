package actuator

import (
	"context"
	"fmt"
	"os/exec"
)

type DirectShutdown struct{}

func NewDirectShutdown() *DirectShutdown {
	return &DirectShutdown{}
}

func (a *DirectShutdown) Name() string {
	return "DirectShutdown"
}

func (a *DirectShutdown) Trigger(ctx context.Context) error {
	// Tries standard linux shutdown command
	cmd := exec.CommandContext(ctx, "shutdown", "-h", "now")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("direct shutdown failed: %v, output: %s", err, string(output))
	}
	
	return nil
}
