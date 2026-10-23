// SPDX-License-Identifier: Apache-2.0 OR MIT
// Copyright (c) 2026 Sebastien Rousseau

// Package collect defines how dot-top measures the host and schedules the
// measurements. A Collector reads one source (CPU, memory, processes, agent
// transcripts) into a caller-owned Sample; the Scheduler runs each collector
// on its own goroutine and publishes copies of the results.
package collect

import (
	"context"
	"time"
)

// Collector measures one source on a fixed interval.
type Collector interface {
	Name() string
	Interval() time.Duration
	Collect(ctx context.Context, dst *Sample) error // reuse buffers; never retain dst
	Caps() Capabilities                             // what this host supports, with a reason for each gap
}

// Metric is one measured value. Only measured values are ever added: a value
// the host cannot provide is reported through Capabilities, never as a
// placeholder number.
type Metric struct {
	Name  string
	Unit  string
	Value float64
}

// Sample is the output of one Collect call. The Scheduler owns the Sample it
// passes to Collect and reuses it on the next tick.
type Sample struct {
	Time    time.Time
	Metrics []Metric
}

// Reset empties s while keeping its buffers.
func (s *Sample) Reset() {
	s.Time = time.Time{}
	s.Metrics = s.Metrics[:0]
}

// Add appends a measured metric.
func (s *Sample) Add(name, unit string, v float64) {
	s.Metrics = append(s.Metrics, Metric{Name: name, Unit: unit, Value: v})
}

// Clone returns a deep copy that shares no memory with s.
func (s *Sample) Clone() Sample {
	c := Sample{Time: s.Time}
	if len(s.Metrics) > 0 {
		c.Metrics = append([]Metric(nil), s.Metrics...)
	}
	return c
}

// Capability says whether one metric is available on this host. When
// Supported is false, Reason says why in words a user can act on
// (for example "needs /proc; not mounted").
type Capability struct {
	Metric    string `json:"metric"`
	Supported bool   `json:"supported"`
	Reason    string `json:"reason,omitempty"`
}

// Capabilities lists one Capability per metric a collector knows about.
type Capabilities []Capability

// Lookup returns the capability for metric and whether the collector
// declared it at all.
func (c Capabilities) Lookup(metric string) (Capability, bool) {
	for _, x := range c {
		if x.Metric == metric {
			return x, true
		}
	}
	return Capability{}, false
}
