//
// Copyright 2018-present Sonatype Inc.
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

// Package iq has definitions and functions for processing golang purls with Nexus IQ Server
package iq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	hashtypes "github.com/sonatype-nexus-community/hashbrowns/types"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	"github.com/sonatype-nexus-community/nancy/types"
	"github.com/sonatype-nexus-community/nancy/useragent"
)

const internalApplicationIDURL = "/api/v2/applications?publicId="

const thirdPartyAPILeft = "/api/v2/scan/applications/"

const thirdPartyAPIRight = "/sources/nancy?stageId="

const (
	pollInterval = 1 * time.Second
)

var (
	localConfig *hashtypes.Config
	tries       = 0
)

// Internal types for use by this package, don't need to expose them
type applicationResponse struct {
	Applications []application `json:"applications"`
}

type application struct {
	ID string `json:"id"`
}

type thirdPartyAPIResult struct {
	StatusURL string `json:"statusUrl"`
}

var statusURLResp types.StatusURLResult

// AuditPackages accepts a slice of purls, public application ID, and configuration, and will submit these to
// Nexus IQ Server for audit, and return a struct of StatusURLResult
func AuditPackages(sbom string, config *hashtypes.Config) (types.StatusURLResult, error) {
	localConfig = config

	if localConfig.User == "admin" && localConfig.Token == "admin123" {
		warnUserOfBadLifeChoices()
	}

	internalID, err := getInternalApplicationID(config.Application)
	if internalID == "" && err != nil {
		return statusURLResp, err
	}

	statusURL, err := submitToThirdPartyAPI(sbom, internalID)
	if statusURL == "" || err != nil {
		return statusURLResp, fmt.Errorf("There was an issue submitting your sbom to the Nexus IQ Third Party API, sbom: %s", sbom)
	}

	statusURLResp = types.StatusURLResult{}

	finished := make(chan bool)

	go func() {
		for {
			select {
			case <-finished:
				return
			default:
				pollIQServer(fmt.Sprintf("%s/%s", localConfig.Server, statusURL), finished, localConfig.MaxRetries)
				time.Sleep(pollInterval)
			}
		}
	}()

	<-finished
	return statusURLResp, nil
}

func getInternalApplicationID(applicationID string) (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s%s%s", localConfig.Server, internalApplicationIDURL, applicationID),
		nil,
	)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(localConfig.User, localConfig.Token)
	req.Header.Set("User-Agent", useragent.GetUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		var response applicationResponse
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			return "", err
		}

		if response.Applications != nil && len(response.Applications) > 0 {
			return response.Applications[0].ID, nil
		}

		return "", fmt.Errorf("Unable to retrieve an internal ID for the specified public application ID: %s", applicationID)
	}

	return "", fmt.Errorf("Unable to communicate with Nexus IQ Server, status code returned is: %d", resp.StatusCode)
}

func submitToThirdPartyAPI(sbom string, internalID string) (string, error) {
	client := &http.Client{}

	url := fmt.Sprintf("%s%s", localConfig.Server, fmt.Sprintf("%s%s%s%s", thirdPartyAPILeft, internalID, thirdPartyAPIRight, localConfig.Stage))

	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer([]byte(sbom)),
	)

	if err != nil {
		return "", err
	}

	req.SetBasicAuth(localConfig.User, localConfig.Token)
	req.Header.Set("User-Agent", useragent.GetUserAgent())
	req.Header.Set("Content-Type", "application/xml")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	// by, _ := ioutil.ReadAll(resp.Body)

	// fmt.Print(string(by))

	if resp.StatusCode == http.StatusAccepted {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		var response thirdPartyAPIResult
		err = json.Unmarshal(bodyBytes, &response)
		return response.StatusURL, nil
	}

	return "", nil
}

func pollIQServer(statusURL string, finished chan bool, maxRetries int) {
	if tries > maxRetries {
		finished <- true
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", statusURL, nil)
	customerrors.Check(err, "Could not poll iQ server")

	req.SetBasicAuth(localConfig.User, localConfig.Token)

	req.Header.Set("User-Agent", useragent.GetUserAgent())

	resp, err := client.Do(req)

	if err != nil {
		finished <- true
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		var response types.StatusURLResult
		err = json.Unmarshal(bodyBytes, &response)

		statusURLResp = response
		if response.IsError {
			finished <- true
		}
		finished <- true
	}
	tries++
	fmt.Print(".")
}

func warnUserOfBadLifeChoices() {
	fmt.Println()
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	fmt.Println("!!!! WARNING : You are using the default username and password for Nexus IQ. !!!!")
	fmt.Println("!!!! You are strongly encouraged to change these, and use a token.           !!!!")
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	fmt.Println()
}
