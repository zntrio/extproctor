// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package cli

import (
	"github.com/spf13/cobra"
)

var (
	// Global flags
	target     string
	unixSocket string
	tlsEnable  bool
	tlsCert    string
	tlsKey     string
	tlsCA      string
	parallel   int
	output     string
	verbose    bool
	filter     string
	tags       []string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "extproctor",
	Short: "A test runner for Envoy ExtProc implementations",
	Long: `ExtProctor is a Go-based test runner designed for validating Envoy External 
Processing (ExtProc) filter implementations. It reads test manifests defined 
using protobuf messages encoded in Prototext and validates that a given ExtProc 
service behaves as expected.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Connection flags
	rootCmd.PersistentFlags().StringVar(&target, "target", "localhost:50051", "ExtProc service address (host:port)")
	rootCmd.PersistentFlags().StringVar(&unixSocket, "unix-socket", "", "Unix domain socket path for ExtProc service (alternative to --target)")
	rootCmd.PersistentFlags().BoolVar(&tlsEnable, "tls", false, "Enable TLS for gRPC connection")
	rootCmd.PersistentFlags().StringVar(&tlsCert, "tls-cert", "", "TLS client certificate file")
	rootCmd.PersistentFlags().StringVar(&tlsKey, "tls-key", "", "TLS client key file")
	rootCmd.PersistentFlags().StringVar(&tlsCA, "tls-ca", "", "TLS CA certificate file")

	// Mark target and unix-socket as mutually exclusive
	rootCmd.MarkFlagsMutuallyExclusive("target", "unix-socket")

	// Execution flags
	rootCmd.PersistentFlags().IntVarP(&parallel, "parallel", "p", 1, "Number of parallel test executions")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "human", "Output format (human, json)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Filtering flags
	rootCmd.PersistentFlags().StringVar(&filter, "filter", "", "Filter tests by name pattern")
	rootCmd.PersistentFlags().StringSliceVar(&tags, "tags", nil, "Filter tests by tags (comma-separated)")
}
