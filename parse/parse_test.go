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
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSha1File(t *testing.T) {
	results, err := Sha1File(path.Join("testdata", "thing.txt"))

	assert.Nil(t, err)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, "main.go", results[0].Location)
	assert.Equal(t, "9987ca4f73d5ea0e534dfbf19238552df4de507e", results[0].Sha1)
	assert.Equal(t, "Makefile", results[1].Location)
	assert.Equal(t, "2a72a07fbc9de22308d12a32f7d33504349e63c9", results[1].Sha1)
}

func TestParseSha1FileBadPath(t *testing.T) {
	results, err := Sha1File(path.Join("testdata", "doesnotexist.txt"))

	assert.Nil(t, results)
	assert.NotNil(t, err)
	typeOfError := reflect.TypeOf(err).String()
	// typeOfError changes with newer golang versions.
	// as of golang 1.16, will be *fs.PathError
	// older versions would yield *os.PathError
	assert.True(t, strings.HasSuffix(typeOfError, "s.PathError"))
}
