package vigil

import "context"

// Checker is an interface for components that check system activity.
// It returns true if the system is considered active (and should NOT sleep).
type Checker interface {
	Name() string
	IsActive(ctx context.Context) (bool, error)
}

// Actuator is an interface for components that perform the actual power state change.
type Actuator interface {
	Name() string
	Trigger(ctx context.Context) error
}

// Reporter is an interface for sending heartbeat/status to the controller.
type Reporter interface {
	Report(ctx context.Context, load float64) (string, error)
}
