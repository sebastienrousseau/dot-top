<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->
<!-- Copyright (c) 2026 Sebastien Rousseau -->

# Roadmap

What dot-top intends to do, and not do, over the next year. Phases follow
the dot-top plan (T1 to T4). Each phase ships in the weekly dot family
release that completes it, at the family version.

## T1: foundation (now)

- Collector framework, scheduler and store (done: store, scheduler, braille
  graph).
- CPU, memory, disk and network collectors on Linux and macOS, checked
  against `ps`, `vm_stat` and `netstat`.
- Process collector (Linux `/proc` dirfd, macOS libproc through purego).
- Layout engine, themes (including btop theme import and the dot palette).
- Budgets as release gates: at most 1% of one core and 20 MiB at a 1 s
  interval with 600 processes.
- First signed release, installable with `mise use github:sebastienrousseau/dot-top`.

## T2: processes, agents and control

- Process table with sort, filter, search, tree and a detail pane.
- Signals, renice and suspend/resume, each identity-checked against PID
  reuse.
- Agents box: Claude Code and Codex sessions attributed to their process
  trees, with tokens, cache hit rate and outcome proxies.
- Hook events through `dot cli hook`; replaces `dot-ai-tui` in the dotfiles.

## T3: hardware depth and history

- GPU (NVML, amdgpu, Intel, Apple IOReport), power, sensors, battery,
  pressure stall and cgroup v2 views.
- Recording and replay, reports by agent, model and project, CSV and JSON
  export.
- Rate limits and latency from the dot-ai gateway; budgets and alerts.

## T4: reach (each item needs a user story first)

- Optional Linux eBPF accounting, remote hosts over ssh stdio, OTLP export,
  local energy rules.

## What dot-top will not do

- Install, configure or govern agents (that is dot-ai and dot-agent).
- Check the health of the dotfiles configuration (dot-doctor).
- Send telemetry, open network listeners, or gate any local safety feature.
- Invent numbers: anything not measured shows "n/a" with the reason.
- Add a paid tier: dot-top stays Apache-2.0 OR MIT.
