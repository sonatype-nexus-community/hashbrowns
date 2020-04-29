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
	"errors"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/hashbrowns/iq"
	"github.com/sonatype-nexus-community/hashbrowns/logger"
	"github.com/sonatype-nexus-community/hashbrowns/parse"
	"github.com/sonatype-nexus-community/hashbrowns/types"

	"github.com/sonatype-nexus-community/nancy/cyclonedx"
	"github.com/spf13/cobra"
)

var log *logrus.Logger

// fryCmd represents the fry command
var fryCmd = &cobra.Command{
	Use:   "fry",
	Short: "Submit list of sha1s to Nexus IQ Server",
	Long: `Provided a path to a file with sha1's and locations, this command will submit them to Nexus IQ Server.

This can be used to audit generic environments for matches to known hashes that do not meet your org's policy.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			if r := recover(); r != nil {
				var ok bool
				err, ok = r.(error)
				if !ok {
					err = fmt.Errorf("pkg: %v", r)
				}

				logger.PrintErrorAndLogLocation(r)
			}
		}()

		log = logger.GetLogger("", config.LogLevel)

		log.Info("Running Fry Command")

		var exitCode int
		if exitCode, err = doParseSha1List(&config); err != nil {
			return
		} else {
			log.WithField("exit_code", exitCode).Trace("Obtained an exit code, exiting")
			// TODO use something like ErrorExit custom error to pass up exit code, instead of calling os.Exit() here
			os.Exit(exitCode)
		}
		return
	},
}

func init() {
	rootCmd.AddCommand(fryCmd)

	fryCmd.PersistentFlags().StringVar(&config.Path, "path", "", "Path to file with sha1s")
	fryCmd.PersistentFlags().StringVar(&config.User, "user", "admin", "Specify Nexus IQ username for request")
	fryCmd.PersistentFlags().StringVar(&config.Token, "token", "admin123", "Specify Nexus IQ token/password for request")
	fryCmd.PersistentFlags().StringVar(&config.Server, "server-url", "http://localhost:8070", "Specify Nexus IQ Server URL")
	fryCmd.PersistentFlags().StringVar(&config.Application, "application", "", "Specify application ID for request")
	fryCmd.PersistentFlags().StringVar(&config.Stage, "stage", "develop", "Specify stage for application")
	fryCmd.PersistentFlags().IntVar(&config.MaxRetries, "max-retries", 300, "Specify maximum number of tries to poll Nexus IQ Server")
}

func doParseSha1List(config *types.Config) (exitCode int, err error) {
	log.WithField("path", config.Path).Info("Checking for existence of path to sha1 file")
	if _, err = os.Stat(config.Path); os.IsNotExist(err) {
		log.WithField("error", err).Error("Path does not exist, returning")

		panic(err)
	}

	log.WithField("path", config.Path).Info("Beginning parsing of file into sha1 type")
	sha1s, err := parse.ParseSha1File(config.Path)
	if err != nil {
		log.WithField("error", err).Error("Error parsing sha1 file into sha1 type")

		panic(err)
	}
	log.WithField("sha1s", sha1s).Debug("Obtained sha1 struct from ParseSha1File")

	log.WithField("sha1s", sha1s).Info("Beginning to obtain SBOM")
	sbom := cyclonedx.SBOMFromSHA1(sha1s)
	log.WithField("sbom", sbom).Trace("SBOM obtained")

	log.WithField("sbom", sbom).Info("Beginning to submit SBOM to Nexus IQ Server")
	res, err := iq.AuditPackages(sbom, config)
	if err != nil {
		log.WithField("error", err).Error("Unable to submit SBOM to Nexus IQ Server")

		panic(err)
	}
	log.WithField("res", res).Trace("Obtained response from Nexus IQ Server")

	fmt.Println()
	if res.IsError {
		log.WithField("err", res.ErrorMessage).Error("Nexus IQ Server responded with an error")
		return 2, errors.New(res.ErrorMessage)
	}

	if res.PolicyAction != "Failure" {
		log.WithField("policy_action", res.PolicyAction).Trace("Nexus IQ Server policy evaluation returned a Failure Policy Action")
		fmt.Println("Wonderbar! No policy violations reported for this audit!")
		fmt.Println("Report URL: ", res.ReportHTMLURL)
		return
	} else {
		log.WithField("policy_action", res.PolicyAction).Trace("Nexus IQ Server policy evaluation returned policy results")
		fmt.Println("Hi, Hashbrowns here, you have some policy violations to clean up!")
		fmt.Println("Report URL: ", res.ReportHTMLURL)
		return 1, nil
	}
}
