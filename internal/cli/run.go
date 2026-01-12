// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"zntr.io/extproctor/internal/client"
	"zntr.io/extproctor/internal/manifest"
	"zntr.io/extproctor/internal/reporter"
	"zntr.io/extproctor/internal/runner"
)

var updateGolden bool

var runCmd = &cobra.Command{
	Use:   "run [paths...]",
	Short: "Run ExtProc tests from manifest files",
	Long: `Run executes ExtProc tests defined in prototext manifest files against 
a target ExtProc service. Multiple paths (files or directories) can be specified.

Examples:
  # Run all tests in a directory
  extproctor run ./tests/ --target localhost:50051

  # Run with Unix domain socket
  extproctor run ./tests/ --unix-socket /var/run/extproc.sock

  # Run with filtering and parallel execution  
  extproctor run ./tests/ --target localhost:50051 --filter "auth*" --parallel 4

  # JSON output for CI
  extproctor run ./tests/ --target localhost:50051 --output json

  # Update golden files
  extproctor run ./tests/ --target localhost:50051 --update-golden`,
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
	RunE:         runTests,
}

func init() {
	runCmd.Flags().BoolVar(&updateGolden, "update-golden", false, "Update golden files with actual responses")
	rootCmd.AddCommand(runCmd)
}

func runTests(cmd *cobra.Command, args []string) error {
	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	// Load manifests from paths
	loader := manifest.NewLoader()
	manifests, err := loader.LoadPaths(args)
	if err != nil {
		return fmt.Errorf("failed to load manifests: %w", err)
	}

	if len(manifests) == 0 {
		return fmt.Errorf("no test manifests found in specified paths")
	}

	// Create reporter based on output format
	var rep reporter.Reporter
	switch output {
	case "json":
		rep = reporter.NewJSONReporter(os.Stdout)
	default:
		rep = reporter.NewHumanReporter(os.Stdout, verbose)
	}

	// Create ExtProc client
	var clientOpts []client.Option
	if unixSocket != "" {
		clientOpts = append(clientOpts, client.WithUnixSocket(unixSocket))
	} else {
		clientOpts = append(clientOpts, client.WithTarget(target))
		if tlsEnable {
			clientOpts = append(clientOpts, client.WithTLS(tlsCert, tlsKey, tlsCA))
		}
	}
	extProcClient, err := client.New(clientOpts...)
	if err != nil {
		return fmt.Errorf("failed to create ExtProc client: %w", err)
	}
	defer func() { _ = extProcClient.Close() }()

	// Create and configure runner
	runnerOpts := []runner.Option{
		runner.WithParallel(parallel),
		runner.WithReporter(rep),
		runner.WithVerbose(verbose),
	}
	if filter != "" {
		runnerOpts = append(runnerOpts, runner.WithFilter(filter))
	}
	if len(tags) > 0 {
		runnerOpts = append(runnerOpts, runner.WithTags(tags))
	}
	if updateGolden {
		runnerOpts = append(runnerOpts, runner.WithUpdateGolden(true))
	}

	testRunner := runner.New(extProcClient, runnerOpts...)

	// Run tests
	results, err := testRunner.Run(ctx, manifests)
	if err != nil {
		return fmt.Errorf("test execution failed: %w", err)
	}

	// Check for failures
	if results.Failed > 0 {
		return fmt.Errorf("%d test(s) failed", results.Failed)
	}

	return nil
}
