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
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/go-sona-types/cyclonedx"
	"github.com/sonatype-nexus-community/hashbrowns/iq"
	"github.com/sonatype-nexus-community/hashbrowns/logger"
	"github.com/sonatype-nexus-community/hashbrowns/parse"
	"github.com/sonatype-nexus-community/hashbrowns/types"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var log *logrus.Logger

var sbomCreator *cyclonedx.CycloneDX

// fryCmd represents the fry command
var fryCmd = &cobra.Command{
	Use:   "fry",
	Short: "Submit list of sha1s to Nexus IQ Server",
	Long: `Provided a path to a file with sha1's and locations, this command will submit them to Nexus IQ Server.

This can be used to audit generic environments for matches to known hashes that do not meet your org's policy.`,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		defer func() {
			if r := recover(); r != nil {
				var ok bool
				err, ok = r.(error)
				if !ok {
					err = fmt.Errorf("pkg: %v", r)
				}

				logger.PrintErrorAndLogLocation(err)
			}
		}()

		fflags := cmd.Flags()

		checkRequiredFlags(fflags)

		log = logger.GetLogger("", config.LogLevel)

		sbomCreator = cyclonedx.Default(log)

		log.Info("Running Fry Command")

		sha1s, err := doParseSha1List(&config)
		if err != nil {
			panic(err)
		}

		var exitCode int
		if exitCode, err = doCycloneDxAndIQ(sha1s); err != nil {
			panic(err)
		}

		if exitCode == 0 {
			return
		}

		return fmt.Errorf("Non zero exit code: %d", exitCode)
	},
}

func init() {
	rootCmd.AddCommand(fryCmd)

	pf := fryCmd.PersistentFlags()

	pf.StringVar(&config.Path, "path", "", "Path to file with sha1s (required)")
	pf.StringVar(&config.User, "user", "admin", "Specify Nexus IQ username for request")
	pf.StringVar(&config.Token, "token", "admin123", "Specify Nexus IQ token/password for request")
	pf.StringVar(&config.Server, "server-url", "http://localhost:8070", "Specify Nexus IQ Server URL")
	pf.StringVar(&config.Application, "application", "", "Specify application ID for request (required)")
	pf.StringVar(&config.Stage, "stage", "develop", "Specify stage for application")
	pf.IntVar(&config.MaxRetries, "max-retries", 300, "Specify maximum number of tries to poll Nexus IQ Server")
}

func checkRequiredFlags(flags *pflag.FlagSet) {
	if !flags.Changed("path") {
		panic(fmt.Errorf("Path not set, see usage for more information"))
	}
	if !flags.Changed("application") {
		panic(fmt.Errorf("Application not set, see usage for more information"))
	}
}

func doParseSha1List(config *types.Config) (sha1s []cyclonedx.Sha1SBOM, err error) {
	log.WithField("path", config.Path).Info("Checking for existence of path to sha1 file")
	if _, err = os.Stat(config.Path); os.IsNotExist(err) {
		log.WithField("error", err).Error("Path does not exist, returning")

		return
	}

	log.WithField("path", config.Path).Info("Beginning parsing of file into sha1 type")
	sha1s, err = parse.Sha1File(config.Path)
	if err != nil {
		log.WithField("error", err).Error("Error parsing sha1 file into sha1 type")

		return
	}
	log.WithField("sha1s", sha1s).Debug("Obtained sha1 struct from ParseSha1File")

	return
}

func doCycloneDxAndIQ(sha1s []cyclonedx.Sha1SBOM) (exitCode int, err error) {
	log.WithField("sha1s", sha1s).Info("Beginning to obtain SBOM")
	sbom := sbomCreator.FromSHA1s(sha1s)
	log.Info("Removing newlines from sbom")
	sbom = strings.Replace(sbom, "\n", "", -1)

	log.WithField("sbom", sbom).Trace("SBOM obtained")

	log.WithField("sbom", sbom).Info("Beginning to submit SBOM to Nexus IQ Server")
	res, err := iq.AuditPackages(sbom, &config)
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
		log.WithField("policy_action", res.PolicyAction).Trace("Nexus IQ Server policy evaluation returned policy results")
		fmt.Println("Wonderbar! No policy violations reported for this audit!")
		fmt.Println("Report URL: ", res.ReportHTMLURL)
		return 0, nil
	}
	log.WithField("policy_action", res.PolicyAction).Trace("Nexus IQ Server policy evaluation returned a Failure Policy Action")
	fmt.Println("Hi, Hashbrowns here, you have some policy violations to clean up!")
	fmt.Println("Report URL: ", res.ReportHTMLURL)
	return 1, nil
}
