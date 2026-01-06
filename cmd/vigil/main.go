package main

import (
	"fmt"
	"os"

	"vigil/internal/commands"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "vigil",
		Short: "Vigil: Unraid Energy Saver",
		Long:  `Vigil is a power management system for your Unraid server. It includes both the Agent (Client) and Controller (Manager).`,
	}

	cobra.OnInitialize(commands.InitConfig)

	rootCmd.AddCommand(commands.NewAgentCommand())
	rootCmd.AddCommand(commands.NewControllerCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
