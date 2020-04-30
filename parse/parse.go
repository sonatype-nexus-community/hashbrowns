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
package parse

import (
	"bufio"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/hashbrowns/logger"
	"github.com/sonatype-nexus-community/nancy/types"
)

var log *logrus.Logger

// Sha1File accepts a path to a file that has shasums for files, and returns them as a
// slice of types.Sha1SBOM, or an error if there was an issue processing the file
func Sha1File(path string) (sha1s []types.Sha1SBOM, err error) {
	log = logger.GetLogger("", 0)

	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		sha1s = append(sha1s, parseSpaceSeperatedLocationAndSha1(scanner))
	}

	sha1s = removeDuplicates(sha1s)

	return
}

func parseSpaceSeperatedLocationAndSha1(scanner *bufio.Scanner) (sha1 types.Sha1SBOM) {
	s := strings.Split(scanner.Text(), "  ")
	sha1.Sha1 = s[0]
	sha1.Location = s[1]

	return
}

func removeDuplicates(sha1s []types.Sha1SBOM) (dedupedSha1s []types.Sha1SBOM) {
	log.WithField("sha1s", sha1s).Debug("Beginning to remove duplicates")
	encountered := map[string]bool{}

	for _, v := range sha1s {
		if encountered[v.Sha1] {
			log.WithField("sha1", v).Trace("Found duplicate sha1, eliminating it")
		} else {
			log.WithField("sha1", v).Trace("Unique sha1, adding it")
			encountered[v.Sha1] = true
			dedupedSha1s = append(dedupedSha1s, v)
		}
	}

	log.WithField("sha1s", dedupedSha1s).Debug("Finished removeing duplicates")

	return
}
