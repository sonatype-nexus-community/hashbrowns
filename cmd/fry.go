package cmd

import (
	"errors"
	"fmt"
	"github.com/sonatype-nexus-community/hashbrowns/iq"
	"github.com/sonatype-nexus-community/hashbrowns/parse"
	"github.com/sonatype-nexus-community/hashbrowns/types"
	"os"

	"github.com/sonatype-nexus-community/nancy/cyclonedx"
	"github.com/spf13/cobra"
)

// fryCmd represents the fry command
var fryCmd = &cobra.Command{
	Use:   "fry",
	Short: "Cook up some hashes",
	Long: `Explain how we fry up some hashes, in great detail.

Could also include some description of why you would want to do whatever this thing does.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		fmt.Println("fry called")

		var exitCode int
		if exitCode, err = doParseSha1List(&config); err != nil {
			return
		} else {
			// TODO use something like ErrorExit custom error to pass up exit code, instead of calling os.Exit() here
			os.Exit(exitCode)
		}
		return
	},
}

func init() {
	rootCmd.AddCommand(fryCmd)

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	fryCmd.PersistentFlags().StringVar(&config.Path, "path", "", "Path to file with sha1s")
	fryCmd.PersistentFlags().StringVar(&config.User, "user", "admin", "Specify Nexus IQ username for request")
	fryCmd.PersistentFlags().StringVar(&config.Token, "token", "admin123", "Specify Nexus IQ token/password for request")
	fryCmd.PersistentFlags().StringVar(&config.Server, "server-url", "http://localhost:8070", "Specify Nexus IQ Server URL")
	fryCmd.PersistentFlags().StringVar(&config.Application, "application", "", "Specify application ID for request")
	fryCmd.PersistentFlags().StringVar(&config.Stage, "stage", "develop", "Specify stage for application")
	fryCmd.PersistentFlags().IntVar(&config.MaxRetries, "max-retries", 300, "Specify maximum number of tries to poll Nexus IQ Server")
}

func doParseSha1List(config *types.Config) (exitCode int, err error) {
	if _, err = os.Stat(config.Path); os.IsNotExist(err) {
		return
	}
	sha1s, err := parse.ParseSha1File(config.Path)
	if err != nil {
		return
	}

	sbom := cyclonedx.SBOMFromSHA1(sha1s)

	res, err := iq.AuditPackages(sbom, config)

	fmt.Println()
	if res.IsError {
		return 2, errors.New(res.ErrorMessage)
	}

	if res.PolicyAction != "Failure" {
		fmt.Println("Wonderbar! No policy violations reported for this audit!")
		fmt.Println("Report URL: ", res.ReportHTMLURL)
		return
	} else {
		fmt.Println("Hi, Nancy here, you have some policy violations to clean up!")
		fmt.Println("Report URL: ", res.ReportHTMLURL)
		return 1, nil
	}
}
