package parse

import (
	"bufio"
	"fmt"
	"os"

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
		fmt.Println(scanner.Text())
	}
	return
}
