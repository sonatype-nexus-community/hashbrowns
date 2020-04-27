//
// Copyright © 2020-present Sonatype Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sonatype-nexus-community/hashbrowns/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

var cmdDummy = &cobra.Command{
	Use:   "dummy",
	Short: "Dummy test command",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		return
	},
}

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)
	return output, err
}

func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}

func validateConfigLogging(t *testing.T, expectedOutput string, expectedConfig types.Config, args ...string) {
	rootCmd.AddCommand(cmdDummy)
	defer func() {
		rootCmd.RemoveCommand(cmdDummy)
	}()

	var testArgs []string
	testArgs = append(testArgs, cmdDummy.Use)
	testArgs = append(testArgs, args...)

	// setup default global config
	origConfig := config
	defer func() {
		config = origConfig
	}()
	config = types.Config{}

	output, err := executeCommand(rootCmd, testArgs...)
	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, output)
	assert.Equal(t, expectedConfig, config)
}

func TestRootCommandUnknownCommand(t *testing.T) {
	output, err := executeCommand(rootCmd, "one", "two")
	assert.NotNil(t, err)
	assert.Equal(t, "Error: unknown command \"one\" for \"hashbrowns\"\nRun 'hashbrowns --help' for usage.\n", output)
}

func TestRootCommandNoArgs(t *testing.T) {
	output, err := executeCommand(rootCmd, "")
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(output, rootCmd.Long))
}

func TestRootCommandLogVerbosity(t *testing.T) {
	validateConfigLogging(t, "", types.Config{}, "")
	validateConfigLogging(t, "", types.Config{LogLevel: 1}, "-v")
	validateConfigLogging(t, "", types.Config{LogLevel: 2}, "-vv")
	validateConfigLogging(t, "", types.Config{LogLevel: 3}, "-vvv")
}
