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
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/sonatype-nexus-community/hashbrowns/iq"
	"github.com/sonatype-nexus-community/hashbrowns/parse"
	"github.com/sonatype-nexus-community/hashbrowns/types"
	"github.com/sonatype-nexus-community/nancy/cyclonedx"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "fry" {
		config, err := parseCommandLineArgs(os.Args[2:])
		if err != nil {
			flag.Usage()
			os.Exit(1)
		}
		doParseSha1List(&config)
	} else {
		_, _ = parseCommandLineArgs(os.Args)
		flag.Usage()

		os.Exit(1)
	}
}

func doParseSha1List(config *types.Config) {
	if _, err := os.Stat(config.Path); os.IsNotExist(err) {
		panic(err)
	}
	sha1s, err := parse.ParseSha1File(config.Path)
	if err != nil {
		panic(err)
	}

	sbom := cyclonedx.SBOMFromSHA1(sha1s)

	res, err := iq.AuditPackages(sbom, config)

	fmt.Println()
	if res.IsError {
		panic(errors.New(res.ErrorMessage))
	}

	if res.PolicyAction != "Failure" {
		fmt.Println("Wonderbar! No policy violations reported for this audit!")
		fmt.Println("Report URL: ", res.ReportHTMLURL)
		os.Exit(0)
	} else {
		fmt.Println("Hi, Nancy here, you have some policy violations to clean up!")
		fmt.Println("Report URL: ", res.ReportHTMLURL)
		os.Exit(1)
	}
}

func parseCommandLineArgs(args []string) (config types.Config, err error) {
	iqCommand := flag.NewFlagSet("fry", flag.ExitOnError)
	iqCommand.BoolVar(&config.Info, "v", false, "Set log level to Info")
	iqCommand.BoolVar(&config.Debug, "vv", false, "Set log level to Debug")
	iqCommand.BoolVar(&config.Trace, "vvv", false, "Set log level to Trace")
	iqCommand.StringVar(&config.Path, "path", "", "Path to file with sha1s")
	iqCommand.StringVar(&config.User, "user", "admin", "Specify Nexus IQ username for request")
	iqCommand.StringVar(&config.Token, "token", "admin123", "Specify Nexus IQ token/password for request")
	iqCommand.StringVar(&config.Server, "server-url", "http://localhost:8070", "Specify Nexus IQ Server URL/port")
	iqCommand.StringVar(&config.Application, "application", "", "Specify application ID for request")
	iqCommand.StringVar(&config.Stage, "stage", "develop", "Specify stage for application")
	iqCommand.IntVar(&config.MaxRetries, "max-retries", 300, "Specify maximum number of tries to poll Nexus IQ Server")

	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, `Usage:
	hashbrown fry [options]

	Options:
	`)
		iqCommand.PrintDefaults()
	}

	err = iqCommand.Parse(args)
	if err != nil {
		return config, err
	}

	return config, nil
}
