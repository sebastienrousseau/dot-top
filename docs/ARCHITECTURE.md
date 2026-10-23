<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Architecture

dot-top is a pipeline with four stages. Only the first does I/O.

```
collect ──► store ──► derive ──► ui
(reads)    (rings)   (rates)    (pure Update/View)
```

- **collect** (`internal/collect`): each `Collector` reads one source into a
  caller-owned `Sample` and declares, through `Caps()`, which metrics this
  host supports and why the others are missing. The `Scheduler` runs one
  goroutine per collector. A collection that overruns its interval skips
  the next tick instead of queuing; a result the receiver does not take
  within one interval is dropped and counted. Both counts are in `Stats`.
- **store** (`internal/store`): fixed-capacity ring buffers of timestamped
  samples. `Push` allocates nothing, so memory is bounded by configuration,
  not uptime.
- **derive** (not built): turns counters into rates and attributes
  processes to agent sessions.
- **ui** (`internal/ui`): renders. `internal/ui/graph` draws braille graphs
  from cached columns, so a new sample recomputes one column and a frame
  allocates nothing. The Bubble Tea v2 TUI is not built yet.

`cmd/dot-top` wires the stages and owns flags, output and exit codes.

## Rules

1. **CGO off.** Every build is pure Go (`CGO_ENABLED=0`).
2. **No I/O in the UI.** `internal/ui/...` must not import `os`, `net` or
   `syscall`; `internal/archtest` fails the build if it does.
3. **No invented numbers.** Unmeasured values are `n/a` with a reason,
   never a placeholder.
