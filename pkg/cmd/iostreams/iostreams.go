// SPDX-License-Identifier: Apache-2.0
//
// Copyright © 2023 The listen.dev team <engineering@garnet.ai>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package iostreams

import (
	"os"
	"time"

	generic "github.com/cli/cli/pkg/iostreams"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/mattn/go-isatty"
)

const (
	progressTrackerUpdateFrequency = time.Millisecond * 100
)

type IOStreams struct {
	*generic.IOStreams

	progressTracker        progress.Writer
	progressTrackerEnabled bool
}

func isTerminal(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}

func System() *IOStreams {
	stdoutIsTTY := isTerminal(os.Stdout)
	stderrIsTTY := isTerminal(os.Stderr)

	s := generic.System()

	io := &IOStreams{
		IOStreams: s,
	}

	if stdoutIsTTY && stderrIsTTY {
		io.progressTrackerEnabled = true
	}

	return io
}

func (s *IOStreams) StartProgressTracking() {
	if !s.progressTrackerEnabled {
		return
	}
	pw := progress.NewWriter()
	pw.SetOutputWriter(s.Out)
	pw.SetStyle(progress.StyleBlocks)
	pw.Style().Colors = progress.StyleColorsExample
	pw.Style().Visibility.Speed = true

	pw.SetUpdateFrequency(progressTrackerUpdateFrequency)
	pw.SetMessageWidth(24)
	pw.SetTrackerLength(25)

	s.progressTracker = pw

	go pw.Render()
}

func (s *IOStreams) CreateProgressTracker(message string, total int64) *ProgressTracker {
	if s.progressTracker == nil {
		return nil
	}
	tracker := progress.Tracker{Message: message, Total: total, Units: progress.UnitsDefault}
	s.progressTracker.AppendTracker(&tracker)

	return &ProgressTracker{tracker: &tracker}
}

func (s *IOStreams) LogProgress(message string) {
	if s.progressTracker == nil {
		return
	}
	s.progressTracker.Log(message)
}

func (s *IOStreams) StopProgressTracking() {
	if s.progressTracker == nil {
		return
	}
	s.progressTracker.Stop()
	s.progressTracker = nil
}

type ProgressTracker struct {
	tracker *progress.Tracker
}

func (pt *ProgressTracker) Increment(value int64) {
	if pt == nil {
		return
	}
	if pt.tracker == nil {
		return
	}
	pt.tracker.Increment(value)
}

func (pt *ProgressTracker) IncrementWithError(value int64) {
	if pt == nil {
		return
	}
	if pt.tracker == nil {
		return
	}
	pt.tracker.IncrementWithError(value)
}

func (pt *ProgressTracker) UpdateMessage(message string) {
	if pt == nil {
		return
	}
	if pt.tracker == nil {
		return
	}
	pt.tracker.UpdateMessage(message)
}
