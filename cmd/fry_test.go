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
	"testing"

	"github.com/sonatype-nexus-community/hashbrowns/types"
	"github.com/stretchr/testify/assert"
)

func validateConfigFryError(t *testing.T, expectedErrorMsgSnippet string, expectedConfig types.Config, args ...string) {
	_, err := executeCommand(rootCmd, args...)

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), expectedErrorMsgSnippet)
}

func TestFryCommandConfigDefaultsIncomplete(t *testing.T) {
	validateConfigFryError(t,
		"stat : no such file or directory",
		types.Config{User: "admin", Token: "admin123", Server: "http://localhost:8070", Stage: "develop", MaxRetries: 300},
		"fry")
}

// TODO: Test errors in CircleCI, likely a different error message, we should likely use httpmock to simulate the response so we get
// something more predictable
func TestFryCommandConfigNoServerRunning(t *testing.T) {
	validateConfigFryError(t,
		"Get \"http://localhost:8070/api/v2/applications?publicId=\": dial tcp 127.0.0.1:8070: connect: connection refused",
		types.Config{User: "admin", Token: "admin123", Server: "http://localhost:8070", Stage: "develop", MaxRetries: 300,
			Path: "testdata/emptyFile"},
		"fry", "--path=testdata/emptyFile")
}
