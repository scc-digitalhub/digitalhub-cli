// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"dhcli/handlers/proxy"
	"dhcli/handlers/utils"
	"dhcli/pkg"
	"dhcli/pkg/flags"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
)

var portForwardCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	projectFlag := flags.NewStringFlag("project", "p", "Mandatory", "")
	localPortFlag := flags.NewStringFlag("local-port", "l", "Local port for listening (default: random)", "")

	cmd := &cobra.Command{
		Use:   "port-forward <run-id>",
		Short: "Start local port-forward for a specific run",
		Long:  "Starts a local port-forward that tunnels requests to the service URL resolved from the run resource, through the configured remote proxy with Authorization",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runID := args[0]

			project := utils.ResolveProject(*projectFlag.Value)
			if project == "" {
				log.Fatalf("Project flag is mandatory (use --project flag or set PROJECT_NAME env variable)")
			}

			if err := utils.RegisterIniCfgWithViper(*envFlag.Value); err != nil {
				log.Fatalf("Failed to load configuration: %v", err)
			}

			// Parse local port
			localPort := 0
			if *localPortFlag.Value != "" {
				port, err := strconv.Atoi(*localPortFlag.Value)
				if err != nil || port < 0 || port > 65535 {
					log.Fatalf("Invalid local port: %s", *localPortFlag.Value)
				}
				localPort = port
			}

			// Create a context that can be cancelled by signals
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Handle shutdown signals
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				sig := <-sigCh
				utils.GetGlobalLogger().Info("Received signal: " + sig.String())
				cancel()
			}()

			// Start the port-forward
			if err := proxy.StartPortForward(ctx, project, runID, localPort); err != nil {
				// Graceful shutdown returns http.ErrServerClosed - this is expected
				if !errors.Is(err, http.ErrServerClosed) {
					log.Fatalf("Port-forward error: %v", err)
				}
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &projectFlag)
	flags.AddFlag(cmd, &localPortFlag)

	return cmd
}()

func init() {
	pkg.RegisterCommand(portForwardCmd)
}
