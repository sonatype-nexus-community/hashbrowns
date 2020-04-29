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

	"github.com/jarcoal/httpmock"
	"github.com/sonatype-nexus-community/hashbrowns/types"
	"github.com/stretchr/testify/assert"
)

const applicationsResponse = `{
	"applications": [
		{
			"id": "4bb67dcfc86344e3a483832f8c496419",
			"publicId": "testapp",
			"name": "TestApp",
			"organizationId": "bb41817bd3e2403a8a52fe8bcd8fe25a",
			"contactUserName": "NewAppContact",
			"applicationTags": [
				{
					"id": "9beee80c6fc148dfa51e8b0359ee4d4e",
					"tagId": "cfea8fa79df64283bd64e5b6b624ba48",
					"applicationId": "4bb67dcfc86344e3a483832f8c496419"
				}
			]
		}
	]
}`

const thirdPartyAPIResultJSON = `{
		"statusUrl": "api/v2/scan/applications/4bb67dcfc86344e3a483832f8c496419/status/9cee2b6366fc4d328edc318eae46b2cb"
}`

const pollingResult = `{
	"policyAction": "None",
	"reportHtmlUrl": "http://sillyplace.com:8090/ui/links/application/test-app/report/95c4c14e",
	"isError": false
}`

func validateConfigFryError(t *testing.T, expectedErrorMsgSnippet string, expectedConfig types.Config, args ...string) {
	_, err := executeCommand(rootCmd, args...)

	assert.NotNil(t, err)
	assert.Equal(t, expectedErrorMsgSnippet, err.Error())
}

func TestFryCommandConfigDefaultsMissingPath(t *testing.T) {
	validateConfigFryError(t,
		"Path not set, see usage for more information",
		types.Config{User: "admin", Token: "admin123", Server: "http://localhost:8070", Stage: "develop", MaxRetries: 300},
		"fry")
}

func TestFryCommandConfigDefaultsMissingApplication(t *testing.T) {
	validateConfigFryError(t,
		"Application not set, see usage for more information",
		types.Config{User: "admin", Token: "admin123", Server: "http://localhost:8070", Stage: "develop", MaxRetries: 300, Path: "test/path"},
		"fry", "--path=test/path")
}

func TestFryCommandConfigNoServerRunning(t *testing.T) {
	validateConfigFryError(t,
		"Get \"http://localhost:8070/api/v2/applications?publicId=testapp\": dial tcp 127.0.0.1:8070: connect: connection refused",
		types.Config{User: "admin", Token: "admin123", Server: "http://localhost:8070", Stage: "develop", MaxRetries: 300,
			Path: "testdata/emptyFile", Application: "testapp"},
		"fry", "--path=testdata/emptyFile", "--application=testapp")
}

func TestFryCommandWithRunningIQ(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://sillyplace.com:8090/api/v2/applications?publicId=testapp",
		httpmock.NewStringResponder(200, applicationsResponse))

	httpmock.RegisterResponder("POST", "http://sillyplace.com:8090/api/v2/scan/applications/4bb67dcfc86344e3a483832f8c496419/sources/nancy?stageId=develop",
		httpmock.NewStringResponder(202, thirdPartyAPIResultJSON))

	httpmock.RegisterResponder("GET", "http://sillyplace.com:8090/api/v2/scan/applications/4bb67dcfc86344e3a483832f8c496419/status/9cee2b6366fc4d328edc318eae46b2cb",
		httpmock.NewStringResponder(200, pollingResult))

	_, err := executeCommand(rootCmd, "fry", "--path=testdata/emptyFile", "--application=testapp", "--server-url=http://sillyplace.com:8090")
	assert.Nil(t, err)
}
