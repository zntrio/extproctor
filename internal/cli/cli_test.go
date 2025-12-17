// SPDX-FileCopyrightText: 2025 Thibault NORMAND
// SPDX-License-Identifier: MIT

package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmd_Basic(t *testing.T) {
	assert.NotNil(t, rootCmd)
	assert.Equal(t, "extproctor", rootCmd.Use)
}

func TestRootCmd_HasFlags(t *testing.T) {
	flags := rootCmd.PersistentFlags()
	assert.NotNil(t, flags)

	// Check target flag
	f := flags.Lookup("target")
	assert.NotNil(t, f)
	assert.Equal(t, "localhost:50051", f.DefValue)

	// Check unix-socket flag
	f = flags.Lookup("unix-socket")
	assert.NotNil(t, f)

	// Check TLS flags
	f = flags.Lookup("tls")
	assert.NotNil(t, f)
	f = flags.Lookup("tls-cert")
	assert.NotNil(t, f)
	f = flags.Lookup("tls-key")
	assert.NotNil(t, f)
	f = flags.Lookup("tls-ca")
	assert.NotNil(t, f)

	// Check execution flags
	f = flags.Lookup("parallel")
	assert.NotNil(t, f)
	assert.Equal(t, "1", f.DefValue)

	f = flags.Lookup("output")
	assert.NotNil(t, f)
	assert.Equal(t, "human", f.DefValue)

	f = flags.Lookup("verbose")
	assert.NotNil(t, f)

	// Check filtering flags
	f = flags.Lookup("filter")
	assert.NotNil(t, f)

	f = flags.Lookup("tags")
	assert.NotNil(t, f)
}

func TestRootCmd_LongDescription(t *testing.T) {
	assert.NotEmpty(t, rootCmd.Long)
	assert.Contains(t, rootCmd.Long, "ExtProc")
}

func TestGlobalFlags_Defaults(t *testing.T) {
	// Test default values are set correctly in rootCmd
	flags := rootCmd.PersistentFlags()

	// Get the default values
	targetFlag := flags.Lookup("target")
	require.NotNil(t, targetFlag)
	assert.Equal(t, "localhost:50051", targetFlag.DefValue)

	parallelFlag := flags.Lookup("parallel")
	require.NotNil(t, parallelFlag)
	assert.Equal(t, "1", parallelFlag.DefValue)

	outputFlag := flags.Lookup("output")
	require.NotNil(t, outputFlag)
	assert.Equal(t, "human", outputFlag.DefValue)
}

func TestExecute_Basic(t *testing.T) {
	// Test that Execute returns without error when rootCmd is set up
	// We don't actually run it as it would try to parse command line args
	assert.NotNil(t, rootCmd)
}

func TestExecute_NoArgs(t *testing.T) {
	// Set args to just the binary name (no subcommand)
	oldArgs := os.Args
	os.Args = []string{"extproctor", "--help"}
	defer func() { os.Args = oldArgs }()

	// Execute should not error on help
	err := Execute()
	assert.NoError(t, err)
}
