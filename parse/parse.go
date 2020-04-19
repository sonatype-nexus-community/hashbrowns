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
