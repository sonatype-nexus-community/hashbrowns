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

// Package logger has functions to obtain a logger, and helpers for setting up where the logger writes
package logger

import (
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/sonatype-nexus-community/nancy/types"
)

const DefaultLogFilename = "hashbrowns.combined.log"

var DefaultLogFile = DefaultLogFilename

var logLady *logrus.Logger

// GetLogger will either return the existing logger, or setup a new logger
func GetLogger(loggerPath string, level int) *logrus.Logger {
	if logLady == nil {
		logLevel := getLoggerLevelFromConfig(level)
		setupLogger(loggerPath, &logLevel)
	}
	return logLady
}

// LogFileLocation will return the location on disk of the log file
func LogFileLocation() (result string, err error) {
	result, _ = os.UserHomeDir()
	err = os.MkdirAll(path.Join(result, types.OssIndexDirName), os.ModePerm)
	if err != nil {
		return
	}
	result = path.Join(result, types.OssIndexDirName, DefaultLogFile)
	return
}

func setupLogger(loggerPath string, level *logrus.Level) (err error) {
	logLady = logrus.New()

	if loggerPath != "" {
		DefaultLogFile = loggerPath
	} else {
		DefaultLogFile = DefaultLogFilename
	}

	if level == nil {
		logLady.Level = logrus.ErrorLevel
	} else {
		logLady.Level = *level
	}

	logLady.Formatter = &logrus.JSONFormatter{}

	location, err := LogFileLocation()
	if err != nil {
		return
	}

	file, err := os.OpenFile(location, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return
	}
	logLady.Out = file

	return
}

func getLoggerLevelFromConfig(level int) logrus.Level {
	switch level {
	case 1:
		return logrus.WarnLevel
	case 2:
		return logrus.InfoLevel
	case 3:
		return logrus.DebugLevel
	case 4:
		return logrus.TraceLevel
	default:
		return logrus.ErrorLevel
	}
}
