// Copyright 2020 Security Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package checker includes structs and functions used for running a check.
package checker

import (
	"fmt"
	"math"
)

type (
	// DetailType is the type of details.
	DetailType int
	// FileType is the type of a file.
	FileType int
)

const (
	// MaxResultConfidence implies full certainty about a check result.
	// TODO(#1393): remove after deprecation.
	MaxResultConfidence = 10
	// HalfResultConfidence signifies uncertainty about a check's score.
	// TODO(#1393): remove after deprecation.
	HalfResultConfidence = 5
	// MinResultConfidence signifies no confidence in the check result.
	// TODO(#1393): remove after deprecation.
	MinResultConfidence = 0
	// TODO(#1393): remove after deprecation.
	migrationThresholdPassValue = 8

	// MaxResultScore is the best score that can be given by a check.
	MaxResultScore = 10
	// MinResultScore is the worst score that can be given by a check.
	MinResultScore = 0
	// InconclusiveResultScore is returned when no reliable information can be retrieved by a check.
	InconclusiveResultScore = -1

	// DetailInfo is info-level log.
	DetailInfo DetailType = iota
	// DetailWarn is warn log.
	DetailWarn
	// DetailDebug is debug log.
	DetailDebug

	// FileTypeNone is a default, not defined.
	FileTypeNone FileType = iota
	// FileTypeSource is for source code files.
	FileTypeSource
	// FileTypeBinary is for binary files.
	FileTypeBinary
	// FileTypeText is for text files.
	FileTypeText
	// FileTypeURL for URLs.
	FileTypeURL

	// OffsetDefault is used if we can't determine the offset, for example when referencing a file but not a
	// specific location in the file.
	OffsetDefault = 1
)

// CheckResult captures result from a check run.
// nolint:govet
type CheckResult struct {
	// TODO(#1393): Remove old structure after deprecation.
	Error      error `json:"-"`
	Name       string
	Details    []string
	Confidence int
	Pass       bool

	// UPGRADEv2: New structure. Omitting unchanged Name field
	// for simplicity.
	Version  int           `json:"-"` // Default value of 0 indicates old structure.
	Error2   error         `json:"-"` // Runtime error indicate a filure to run the check.
	Details2 []CheckDetail `json:"-"` // Details of tests and sub-checks
	Score    int           `json:"-"` // {[-1,0...10], -1 = Inconclusive}
	Reason   string        `json:"-"` // A sentence describing the check result (score, etc)
}

// CheckDetail contains information for each detail.
type CheckDetail struct {
	Msg  LogMessage
	Type DetailType // Any of DetailWarn, DetailInfo, DetailDebug.
}

// LogMessage is a structure that encapsulates detail's information.
// This allows updating the definition easily.
// nolint:govet
type LogMessage struct {
	Text    string   // A short string explaining why the detail was recorded/logged.
	Path    string   // Fullpath to the file.
	Type    FileType // Type of file.
	Offset  int      // Offset in the file of Path (line for source/text files).
	Snippet string   // Snippet of code
	// UPGRADEv3: to remove.
	Version int // `3` to indicate the detail was logged using new structure.
}

// CreateProportionalScore creates a proportional score.
func CreateProportionalScore(success, total int) int {
	if total == 0 {
		return 0
	}

	return int(math.Min(float64(MaxResultScore*success/total), float64(MaxResultScore)))
}

// AggregateScores adds up all scores
// and normalizes the result.
// Each score contributes equally.
func AggregateScores(scores ...int) int {
	n := float64(len(scores))
	r := 0
	for _, s := range scores {
		r += s
	}
	return int(math.Floor(float64(r) / n))
}

// AggregateScoresWithWeight adds up all scores
// and normalizes the result.
func AggregateScoresWithWeight(scores map[int]int) int {
	r := 0
	ws := 0
	for s, w := range scores {
		r += s * w
		ws += w
	}
	return int(math.Floor(float64(r) / float64(ws)))
}

// NormalizeReason - placeholder function if we want to update range of scores.
func NormalizeReason(reason string, score int) string {
	return fmt.Sprintf("%v -- score normalized to %d", reason, score)
}

// CreateResultWithScore is used when
// the check runs without runtime errors and we want to assign a
// specific score.
func CreateResultWithScore(name, reason string, score int) CheckResult {
	pass := true
	if score < migrationThresholdPassValue {
		pass = false
	}
	return CheckResult{
		Name: name,
		// Old structure.
		Error:      nil,
		Confidence: MaxResultScore,
		Pass:       pass,
		// New structure.
		Version: 2,
		Error2:  nil,
		Score:   score,
		Reason:  reason,
	}
}

// CreateProportionalScoreResult is used when
// the check runs without runtime errors and we assign a
// proportional score. This may be used if a check contains
// multiple tests and we want to assign a score proportional
// the the number of tests that succeeded.
func CreateProportionalScoreResult(name, reason string, b, t int) CheckResult {
	pass := true
	score := CreateProportionalScore(b, t)
	if score < migrationThresholdPassValue {
		pass = false
	}
	return CheckResult{
		Name: name,
		// Old structure.
		Error:      nil,
		Confidence: MaxResultConfidence,
		Pass:       pass,
		// New structure.
		Version: 2,
		Error2:  nil,
		Score:   score,
		Reason:  NormalizeReason(reason, score),
	}
}

// CreateMaxScoreResult is used when
// the check runs without runtime errors and we can assign a
// maximum score to the result.
func CreateMaxScoreResult(name, reason string) CheckResult {
	return CreateResultWithScore(name, reason, MaxResultScore)
}

// CreateMinScoreResult is used when
// the check runs without runtime errors and we can assign a
// minimum score to the result.
func CreateMinScoreResult(name, reason string) CheckResult {
	return CreateResultWithScore(name, reason, MinResultScore)
}

// CreateInconclusiveResult is used when
// the check runs without runtime errors, but we don't
// have enough evidence to set a score.
func CreateInconclusiveResult(name, reason string) CheckResult {
	return CheckResult{
		Name: name,
		// Old structure.
		Confidence: 0,
		Pass:       false,
		// New structure.
		Version: 2,
		Score:   InconclusiveResultScore,
		Reason:  reason,
	}
}

// CreateRuntimeErrorResult is used when the check fails to run because of a runtime error.
func CreateRuntimeErrorResult(name string, e error) CheckResult {
	return CheckResult{
		Name: name,
		// Old structure.
		Error:      e,
		Confidence: 0,
		Pass:       false,
		// New structure.
		Version: 2,
		Error2:  e,
		Score:   InconclusiveResultScore,
		Reason:  e.Error(), // Note: message already accessible by caller thru `Error`.
	}
}
