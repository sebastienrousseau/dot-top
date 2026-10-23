// SPDX-License-Identifier: Apache-2.0 OR MIT
// Copyright (c) 2026 Sebastien Rousseau

package collect

import (
	"testing"
	"time"
)

func TestSampleResetKeepsBuffer(t *testing.T) {
	var s Sample
	s.Time = time.Unix(1, 0)
	s.Add("cpu", "%", 12.5)
	backing := &s.Metrics[:1][0]
	s.Reset()
	if len(s.Metrics) != 0 || !s.Time.IsZero() {
		t.Fatalf("Reset left %+v", s)
	}
	s.Add("mem", "B", 1)
	if &s.Metrics[0] != backing {
		t.Fatal("Reset dropped the metrics buffer")
	}
}

func TestSampleCloneSharesNothing(t *testing.T) {
	var s Sample
	s.Add("cpu", "%", 1)
	c := s.Clone()
	s.Metrics[0].Value = 99
	if c.Metrics[0].Value != 1 {
		t.Fatalf("clone aliased the source: %v", c.Metrics[0].Value)
	}
	var empty Sample
	if got := empty.Clone(); got.Metrics != nil {
		t.Fatalf("clone of empty sample = %+v, want nil metrics", got)
	}
}

func TestCapabilitiesLookup(t *testing.T) {
	caps := Capabilities{
		{Metric: "cpu", Supported: true},
		{Metric: "gpu", Supported: false, Reason: "no GPU driver interface"},
	}
	got, ok := caps.Lookup("gpu")
	if !ok || got.Supported || got.Reason != "no GPU driver interface" {
		t.Fatalf("Lookup(gpu) = %+v, %v", got, ok)
	}
	if _, ok := caps.Lookup("disk"); ok {
		t.Fatal("Lookup(disk) found an undeclared metric")
	}
}
