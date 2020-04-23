package cmd

import (
	"github.com/sonatype-nexus-community/hashbrowns/types"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func validateConfigFryError(t *testing.T, expectedErrorMsgSnippet string, expectedConfig types.Config, args ...string) {
	// setup default global config
	config = types.Config{}

	_, err := executeCommand(rootCmd, args...)

	if expectedErrorMsgSnippet != "" {
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), expectedErrorMsgSnippet))
	} else {
		assert.NoError(t, err)
	}

	assert.Equal(t, expectedConfig, config)

	// reset default global config
	config = types.Config{}
}

func TestFryCommandConfigDefaultsIncomplete(t *testing.T) {
	validateConfigFryError(t,
		"stat : no such file or directory",
		types.Config{User: "admin", Token: "admin123", Server: "http://localhost:8070", Stage: "develop", MaxRetries: 300},
		"fry")
}

func TestFryCommandConfigNoServerRunning(t *testing.T) {
	validateConfigFryError(t,
		"Get \"http://localhost:8070/api/v2/applications?publicId=\": dial tcp [::1]:8070: connect: connection refused",
		types.Config{User: "admin", Token: "admin123", Server: "http://localhost:8070", Stage: "develop", MaxRetries: 300,
			Path: "testdata/emptyFile"},
		"fry", "--path=testdata/emptyFile")
}
