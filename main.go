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
package main

import (
	"fmt"
	"os"

	"github.com/common-nighthawk/go-figure"
	"github.com/sonatype-nexus-community/hashbrowns/buildversion"
	"github.com/sonatype-nexus-community/hashbrowns/cmd"
)

func main() {
	printHeader(true)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func printHeader(print bool) {
	if print {
		figure.NewFigure("Hashbrowns", "larry3d", true).Print()
		figure.NewFigure("By Sonatype & Friends", "pepper", true).Print()

		fmt.Println("Hashbrowns version: " + buildversion.BuildVersion)
	}
}
