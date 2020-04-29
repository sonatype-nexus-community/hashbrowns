//
// Copyright 2020-present Sonatype Inc.
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

	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/hashbrowns/logger"
	hashtypes "github.com/sonatype-nexus-community/hashbrowns/types"
	"github.com/sonatype-nexus-community/nancy/customerrors"
	"github.com/sonatype-nexus-community/nancy/types"
	useragent "github.com/sonatype-nexus-community/nancy/useragent"
)

const internalApplicationIDURL = "/api/v2/applications?publicId="

const thirdPartyAPILeft = "/api/v2/scan/applications/"

const thirdPartyAPIRight = "/sources/nancy?stageId="

const contentTypeApplicationXML = "application/xml"

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

var log *logrus.Logger

// AuditPackages accepts a slice of purls, public application ID, and configuration, and will submit these to
// Nexus IQ Server for audit, and return a struct of StatusURLResult
func AuditPackages(sbom string, config *hashtypes.Config) (types.StatusURLResult, error) {
	log = logger.GetLogger("", config.LogLevel)

	useragent.CLIENTTOOL = "hashbrowns-client"
	log.WithField("client", useragent.CLIENTTOOL).Trace("Setting the user agent")

	localConfig = config

	if localConfig.User == "admin" && localConfig.Token == "admin123" {
		log.Trace("Warning user of bad life choices, default Nexus IQ Server user and password")
		warnUserOfBadLifeChoices()
	}

	log.WithField("application_id", config.Application).Debug("Getting internal application ID from Nexus IQ Server")
	internalID, err := getInternalApplicationID(config.Application)
	if internalID == "" && err != nil {
		log.WithField("error", err).Error("Unable to obtain internal application ID from Nexus IQ Server")
		return statusURLResp, err
	}

	log.WithFields(logrus.Fields{
		"internal_id": internalID,
		"sbom":        sbom,
	}).Debug("Submitting SBOM to Nexus IQ Server")
	statusURL, err := submitToThirdPartyAPI(sbom, internalID)
	if statusURL == "" || err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
			"sbom":  sbom,
		}).Error("Unable to submit sbom to Nexus IQ Server")

		return statusURLResp, fmt.Errorf("There was an issue submitting your sbom to the Nexus IQ Third Party API, sbom: %s", sbom)
	}
	log.WithField("status_url", statusURL).Trace("Obtained StatusURL from Nexus IQ Server")

	statusURLResp = types.StatusURLResult{}

	finished := make(chan bool)

	go func() {
		for {
			select {
			case <-finished:
				return
			default:
				log.WithField("status_url", statusURL).Trace("Polling Nexus IQ Server for response")
				pollIQServer(fmt.Sprintf("%s/%s", localConfig.Server, statusURL), finished, localConfig.MaxRetries)
				time.Sleep(pollInterval)
			}
		}
	}()

	<-finished
	return statusURLResp, nil
}

func getInternalApplicationID(applicationID string) (string, error) {
	log.WithField("application_id", applicationID).Debug("Beginning to obtain internal application ID from Nexus IQ Server")
	client := &http.Client{}

	url := fmt.Sprintf("%s%s%s", localConfig.Server, internalApplicationIDURL, applicationID)

	log.WithFields(logrus.Fields{
		"url": url,
	}).Trace("Setting up request to Nexus IQ Server for internal application ID")
	req, err := http.NewRequest(
		"GET",
		url,
		nil,
	)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Unable to obtain internal application ID from Nexus IQ Server")

		return "", err
	}

	log.Info("Setting up basic auth, and getting user agent for request to Nexus IQ Server for internal application ID")
	req.SetBasicAuth(localConfig.User, localConfig.Token)
	req.Header.Set("User-Agent", useragent.GetUserAgent())
	log.WithFields(logrus.Fields{
		"user_agent": useragent.GetUserAgent(),
	}).Trace("Set up basic auth and user agent for request to Nexus IQ Server for internal application ID")

	log.Info("Making request to Nexus IQ Server for internal application ID")
	resp, err := client.Do(req)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
			"resp":  resp,
		}).Error("Unable to obtain internal application ID from Nexus IQ Server")

		return "", err
	}

	defer resp.Body.Close()

	log.Info("Checking response from Nexus IQ Server for internal application ID")
	if resp.StatusCode == http.StatusOK {
		log.Info("Response from Nexus IQ Server for internal application ID valid, moving forward")
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.WithFields(logrus.Fields{
				"error": err,
			}).Error("Unable to obtain internal application ID from Nexus IQ Server")

			return "", err
		}
		log.WithField("body_bytes", string(bodyBytes)).Trace("Obtained a response body from Nexus IQ Server for internal application ID")

		log.Info("Attempting to unmarshal response from Nexus IQ Server")
		var response applicationResponse
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			log.WithFields(logrus.Fields{
				"error": err,
			}).Error("Unable to obtain internal application ID from Nexus IQ Server")

			return "", err
		}
		log.WithField("response", response).Trace("Successfully unmarshal'd response from Nexus IQ Server for internal application ID")

		if response.Applications != nil && len(response.Applications) > 0 {
			log.WithField("internal_application_id", response.Applications[0].ID).Trace("Obtained internal application ID, returning")

			return response.Applications[0].ID, nil
		}

		log.Error("Unable to obtain internal application ID from Nexus IQ Server")
		return "", fmt.Errorf("Unable to retrieve an internal ID for the specified public application ID: %s", applicationID)
	}

	log.WithField("status_code", resp.StatusCode).Error("Unable to obtain internal application ID from Nexus IQ Server")
	return "", fmt.Errorf("Unable to communicate with Nexus IQ Server, status code returned is: %d", resp.StatusCode)
}

func submitToThirdPartyAPI(sbom string, internalID string) (string, error) {
	log.WithFields(logrus.Fields{
		"internal_application_id": internalID,
		"sbom":                    sbom,
	}).Debug("Beginning to submit SBOM to Nexus IQ Server")
	client := &http.Client{}

	url := fmt.Sprintf("%s%s", localConfig.Server, fmt.Sprintf("%s%s%s%s", thirdPartyAPILeft, internalID, thirdPartyAPIRight, localConfig.Stage))

	log.WithFields(logrus.Fields{
		"url": url,
	}).Trace("Setting up request to Nexus IQ Server to submit SBOM")
	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer([]byte(sbom)),
	)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Unable to setup POST request to Nexus IQ Server for submitting SBOM")

		return "", err
	}

	log.Info("Setting up basic auth, getting user agent, and setting content type for request to Nexus IQ Server for submitting SBOM")
	req.SetBasicAuth(localConfig.User, localConfig.Token)
	req.Header.Set("User-Agent", useragent.GetUserAgent())
	req.Header.Set("Content-Type", contentTypeApplicationXML)
	log.WithFields(logrus.Fields{
		"user_agent":   useragent.GetUserAgent(),
		"content_type": contentTypeApplicationXML,
	}).Trace("Set up basic auth, user agent and content type for request to Nexus IQ Server for submitting SBOM")

	log.Info("Making request to Nexus IQ Server for internal application ID")
	resp, err := client.Do(req)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Unable to do POST request to Nexus IQ Server for submitting SBOM")

		return "", err
	}

	defer resp.Body.Close()

	log.Info("Checking response from Nexus IQ Server for submitting SBOM")
	if resp.StatusCode == http.StatusAccepted {
		log.Info("Response valid from Nexus IQ Server for submitting SBOM, moving forward")
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.WithFields(logrus.Fields{
				"error": err,
			}).Error("Unable to read response body from Nexus IQ Server for submitting SBOM")

			return "", err
		}
		log.WithField("body_bytes", string(bodyBytes)).Trace("Obtained a response body from Nexus IQ Server for submitting SBOM")

		log.Info("Attempting to unmarshal response from Nexus IQ Server for submitting SBOM")
		var response thirdPartyAPIResult
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			log.WithFields(logrus.Fields{
				"error": err,
			}).Error("Unable to unmarshal response body from Nexus IQ Server for submitting SBOM")

			return "", err
		}
		log.WithField("response", response).Trace("Successfully unmarshal'd response from Nexus IQ Server for submitting SBOM, returning")

		return response.StatusURL, nil
	}

	log.WithField("status_code", resp.StatusCode).Error("Unable to submit SBOM to Nexus IQ Server")
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
		if err != nil {
			panic(err)
		}

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
