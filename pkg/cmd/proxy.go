// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"dhcli/handlers/proxy"
	"dhcli/handlers/utils"
	"dhcli/pkg"
	"dhcli/pkg/flags"

	"github.com/spf13/cobra"
)

var proxyCmd = func() *cobra.Command {
	envFlag := flags.NewStringFlag("env", "e", "environment", "")
	projectFlag := flags.NewStringFlag("project", "p", "Mandatory", "")

	cmd := &cobra.Command{
		Use:   "proxy <run-id>",
		Short: "Open browser with authenticated access to a remote service",
		Long: `Bootstraps an authenticated browser session to the service exposed by the given run.

The CLI starts a temporary localhost HTTP server, opens the browser, and serves
a single page that automatically posts credentials to the remote proxy's /auth
endpoint. The browser then communicates directly with the remote service.

Unlike port-forward, no HTTP traffic is proxied through the CLI. This command
is intended for browser access (WebSockets, SSE, cookies all work naturally).`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runID := args[0]

			project := utils.ResolveProject(*projectFlag.Value)
			if project == "" {
				log.Fatalf("Project flag is mandatory (use --project flag or set PROJECT_NAME env variable)")
			}

			if err := utils.RegisterIniCfgWithViper(*envFlag.Value); err != nil {
				log.Fatalf("Failed to load configuration: %v", err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				sig := <-sigCh
				utils.GetGlobalLogger().Info("Received signal: " + sig.String())
				cancel()
			}()

			if err := proxy.StartBrowserProxy(ctx, project, runID); err != nil {
				log.Fatalf("Proxy error: %v", err)
			}
		},
	}

	flags.AddFlag(cmd, &envFlag)
	flags.AddFlag(cmd, &projectFlag)

	return cmd
}()

func init() {
	pkg.RegisterCommand(proxyCmd)
}
