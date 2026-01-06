package commands

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vigil/internal/agent/actuator"
	"vigil/internal/agent/client"

	"github.com/shirou/gopsutil/v4/load"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewAgentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Run Vigil Agent",
		Run: func(cmd *cobra.Command, args []string) {
			startAgent()
		},
	}

	cmd.Flags().String("controller", "http://localhost:8080", "Vigil Controller URL")
	cmd.Flags().Duration("interval", 1*time.Minute, "Check Interval")
	cmd.Flags().Bool("dry-run", false, "Dry Run Mode")

	viper.BindPFlag("controller_url", cmd.Flags().Lookup("controller"))
	viper.BindPFlag("check_interval", cmd.Flags().Lookup("interval"))
	viper.BindPFlag("dry_run", cmd.Flags().Lookup("dry-run"))

	return cmd
}

func startAgent() {
	controllerURL := viper.GetString("controller_url")
	interval := viper.GetDuration("check_interval")
	dryRun := viper.GetBool("dry_run")

	slog.Info("Starting Vigil Agent", "controller", controllerURL, "dry_run", dryRun)

	webClient := client.NewWebControlClient(controllerURL)
	shutActuator := actuator.NewDirectShutdown()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	ctx := context.Background()

	for {
		select {
		case <-sigChan:
			slog.Info("Stopping Vigil Agent")
			return
		case <-ticker.C:
			// 1. Get current load for reporting (Using Load15 as requested)
			avg, _ := load.AvgWithContext(ctx)
			currentLoad := 0.0
			if avg != nil {
				currentLoad = avg.Load15
			}

			// 2. Report to Controller
			action, err := webClient.Report(ctx, currentLoad)
			if err != nil {
				slog.Error("Failed to contact controller", "error", err)
				continue
			}

			// 3. Act
			if action == "sleep" {
				slog.Info("Controller requested SLEEP")

				if dryRun {
					slog.Info("[DRY RUN] Would execute shutdown")
				} else {
					if err := shutActuator.Trigger(ctx); err != nil {
						slog.Error("Shutdown failed", "error", err)
					}
				}
			}
		}
	}
}
