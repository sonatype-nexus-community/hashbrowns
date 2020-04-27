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

	"github.com/sonatype-nexus-community/nancy/types"
)

func ParseSha1File(path string) (sha1s []types.Sha1SBOM, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		sha1s = append(sha1s, parseSpaceSeperatedLocationAndSha1(scanner))
	}

	return
}

func parseSpaceSeperatedLocationAndSha1(scanner *bufio.Scanner) (sha1 types.Sha1SBOM) {
	s := strings.Split(scanner.Text(), " ")
	sha1.Location = s[0]
	sha1.Sha1 = s[1]

	return
}
