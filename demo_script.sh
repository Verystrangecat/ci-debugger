#!/usr/bin/env bash
# Scripted demo output for ci-debugger v0.4 GIF recording.
# Mimics real tool output without requiring Docker.

RESET="\033[0m"
BOLD="\033[1m"
DIM="\033[2m"
GREEN="\033[32m"
RED="\033[31m"
YELLOW="\033[33m"
CYAN="\033[36m"
BLUE="\033[34m"
MAGENTA="\033[35m"
WHITE="\033[97m"

p() { printf "%b\n" "$*"; }
s() { sleep "$1"; }

# ── helper: print with delay ──────────────────────────────────────────────────
slow() {
  local delay="${1:-0.03}"
  shift
  printf "%b" "$*"
  sleep "$delay"
  printf "\n"
}

bar() { p "${DIM}────────────────────────────────────────────────────────────────${RESET}"; }

# ═══════════════════════════════════════════════════════════════════════════════
# SCENE 1 — scan
# ═══════════════════════════════════════════════════════════════════════════════

p ""
slow 0.02 "${BOLD}${CYAN}\$ ci-debugger scan -W testdata/demo_v04.yml${RESET}"
s 0.4

p ""
p "${BOLD}Scan Results${RESET}  ${DIM}(1 workflow)${RESET}"
bar
s 0.15
p "  ${GREEN}✓${RESET}  testdata/demo_v04.yml — no issues found"
s 0.1
bar
p ""
p "  ${GREEN}0 error(s), 0 warning(s)${RESET}"
p ""
s 0.8

# ═══════════════════════════════════════════════════════════════════════════════
# SCENE 2 — env-report
# ═══════════════════════════════════════════════════════════════════════════════

slow 0.02 "${BOLD}${CYAN}\$ ci-debugger run --env-report -W testdata/demo_v04.yml${RESET}"
s 0.4

p ""
p "${BOLD}Environment Variables Report${RESET}"
bar
printf "  %-42s %-12s %s\n" "Variable" "Status" "Value / Note"
bar
s 0.1

printf "  \033[32m%-42s\033[0m %-12s %s\n" "GITHUB_SHA"        "real"        "a3f4b2c1d9e8f7..."
s 0.05
printf "  \033[32m%-42s\033[0m %-12s %s\n" "GITHUB_REF"        "real"        "refs/heads/main"
s 0.05
printf "  \033[32m%-42s\033[0m %-12s %s\n" "GITHUB_REPOSITORY" "real"        "murataslan1/myapp"
s 0.05
printf "  \033[33m%-42s\033[0m %-12s %s\n" "GITHUB_WORKSPACE"  "stubbed"     "/github/workspace"
s 0.05
printf "  \033[33m%-42s\033[0m %-12s %s\n" "RUNNER_OS"         "stubbed"     "Linux"
s 0.05
printf "  \033[31m%-42s\033[0m %-12s %s\n" "GITHUB_TOKEN"      "unavailable" "injected by GitHub"
s 0.05
printf "  \033[31m%-42s\033[0m %-12s %s\n" "GITHUB_ACTOR"      "unavailable" "GitHub user who triggered"
bar
p ""
p "  ${GREEN}real${RESET}       = from local git   ${YELLOW}stubbed${RESET} = fixed value   ${RED}unavailable${RESET} = GitHub-only"
p ""
s 0.8

# ═══════════════════════════════════════════════════════════════════════════════
# SCENE 3 — full run (expressions + outputs + matrix + deploy)
# ═══════════════════════════════════════════════════════════════════════════════

slow 0.02 "${BOLD}${CYAN}\$ ci-debugger run -v -W testdata/demo_v04.yml${RESET}"
s 0.4

p ""
p "${BOLD}CI Pipeline${RESET}  ${DIM}push${RESET}"
p ""
s 0.3

# ── job: build ────────────────────────────────────────────────────────────────
p "${BOLD}▶ build${RESET}  ${DIM}(ghcr.io/catthehacker/ubuntu:act-latest)${RESET}"
s 0.3

p "  ${DIM}[1/4]${RESET} Checkout"
s 0.15
p "    (actions/checkout: workspace already mounted at /github/workspace)"
s 0.2
p "  ${GREEN}✓ [1/4]${RESET} Checkout  ${DIM}(0.0s)${RESET}"
s 0.2

p "  ${DIM}[2/4]${RESET} Show context"
s 0.3
p "    Repository: murataslan1/myapp"
p "    Branch:     refs/heads/main"
p "    SHA:        a3f4b2c1d9e8f7a6b5c4d3e2f1a0b9c8"
s 0.2
p "  ${GREEN}✓ [2/4]${RESET} Show context  ${DIM}(0.3s)${RESET}"
s 0.2

p "  ${DIM}[3/4]${RESET} Set build metadata"
s 0.3
p "  ${GREEN}✓ [3/4]${RESET} Set build metadata  ${DIM}(0.2s)${RESET}"
s 0.2

p "  ${DIM}[4/4]${RESET} Build"
s 0.3
p "    Building myapp v2.1.0..."
p "    Done."
s 0.2
p "  ${GREEN}✓ [4/4]${RESET} Build  ${DIM}(0.4s)${RESET}"
s 0.3

p ""
p "  ${DIM}Job outputs:${RESET}"
p "    version  → ${CYAN}2.1.0${RESET}"
p "    artifact → ${CYAN}myapp-2.1.0.tar.gz${RESET}"
p ""
s 0.4

# ── job: test (matrix) ────────────────────────────────────────────────────────
p "${DIM}  Matrix: 2 combination(s) for job \"test\"${RESET}"
p ""
s 0.2

p "${BOLD}▶ test (go=1.22)${RESET}  ${DIM}(ghcr.io/catthehacker/ubuntu:act-latest)${RESET}"
s 0.2

p "  ${DIM}[1/3]${RESET} Run tests (Go 1.22)"
s 0.3
p "    Testing with Go 1.22..."
s 0.2
p "  ${GREEN}✓ [1/3]${RESET} Run tests (Go 1.22)  ${DIM}(0.4s)${RESET}"
s 0.15

p "  ${DIM}[2/3]${RESET} Check artifact"
s 0.2
p "    Artifact: myapp-2.1.0.tar.gz"
s 0.15
p "  ${GREEN}✓ [2/3]${RESET} Check artifact  ${DIM}(0.2s)${RESET}"
s 0.15

p "  ${DIM}[3/3]${RESET} Coverage check"
s 0.2
p "    Coverage OK"
s 0.15
p "  ${GREEN}✓ [3/3]${RESET} Coverage check  ${DIM}(0.2s)${RESET}"
p ""
s 0.2

p "${BOLD}▶ test (go=1.23)${RESET}  ${DIM}(ghcr.io/catthehacker/ubuntu:act-latest)${RESET}"
s 0.2

p "  ${DIM}[1/3]${RESET} Run tests (Go 1.23)"
s 0.3
p "    Testing with Go 1.23..."
s 0.2
p "  ${GREEN}✓ [1/3]${RESET} Run tests (Go 1.23)  ${DIM}(0.4s)${RESET}"
s 0.15

p "  ${DIM}[2/3]${RESET} Check artifact"
s 0.2
p "    Artifact: myapp-2.1.0.tar.gz"
s 0.15
p "  ${GREEN}✓ [2/3]${RESET} Check artifact  ${DIM}(0.2s)${RESET}"
s 0.15

p "  ${DIM}[3/3]${RESET} Coverage check"
s 0.2
p "    Coverage OK"
s 0.15
p "  ${GREEN}✓ [3/3]${RESET} Coverage check  ${DIM}(0.2s)${RESET}"
p ""
s 0.4

# ── job: deploy ───────────────────────────────────────────────────────────────
p "${BOLD}▶ deploy${RESET}  ${DIM}(ghcr.io/catthehacker/ubuntu:act-latest)${RESET}"
s 0.3

p "  ${DIM}[1/1]${RESET} Deploy 2.1.0"
s 0.4
p "    Deploying myapp-2.1.0.tar.gz"
p "    Version: 2.1.0"
p "    Deploy complete."
s 0.2
p "  ${GREEN}✓ [1/1]${RESET} Deploy 2.1.0  ${DIM}(0.5s)${RESET}"
p ""
s 0.3

# ── summary ───────────────────────────────────────────────────────────────────
p "${BOLD}${BLUE}╭──────────────────────────────────────────────────────╮${RESET}"
p "${BOLD}${BLUE}│${RESET}  Summary                                             ${BOLD}${BLUE}│${RESET}"
p "${BOLD}${BLUE}│${RESET}                                                      ${BOLD}${BLUE}│${RESET}"
p "${BOLD}${BLUE}│${RESET}  Job: build                                          ${BOLD}${BLUE}│${RESET}"
p "${BOLD}${BLUE}│${RESET}  ${GREEN}✓${RESET} 1  Checkout             passed    0.0s          ${BOLD}${BLUE}│${RESET}"
p "${BOLD}${BLUE}│${RESET}  ${GREEN}✓${RESET} 2  Show context         passed    0.3s          ${BOLD}${BLUE}│${RESET}"
p "${BOLD}${BLUE}│${RESET}  ${GREEN}✓${RESET} 3  Set build metadata   passed    0.2s          ${BOLD}${BLUE}│${RESET}"
p "${BOLD}${BLUE}│${RESET}  ${GREEN}✓${RESET} 4  Build                passed    0.4s          ${BOLD}${BLUE}│${RESET}"
p "${BOLD}${BLUE}│${RESET}                                                      ${BOLD}${BLUE}│${RESET}"
p "${BOLD}${BLUE}│${RESET}  Job: test (go=1.22)  Job: test (go=1.23)           ${BOLD}${BLUE}│${RESET}"
p "${BOLD}${BLUE}│${RESET}  ${GREEN}✓${RESET} all 3 steps passed       ${GREEN}✓${RESET} all 3 steps passed   ${BOLD}${BLUE}│${RESET}"
p "${BOLD}${BLUE}│${RESET}                                                      ${BOLD}${BLUE}│${RESET}"
p "${BOLD}${BLUE}│${RESET}  Job: deploy                                         ${BOLD}${BLUE}│${RESET}"
p "${BOLD}${BLUE}│${RESET}  ${GREEN}✓${RESET} 1  Deploy 2.1.0         passed    0.5s          ${BOLD}${BLUE}│${RESET}"
p "${BOLD}${BLUE}│${RESET}                                                      ${BOLD}${BLUE}│${RESET}"
p "${BOLD}${BLUE}│${RESET}  Total: 3.2s   ${GREEN}11 passed${RESET}   0 failed               ${BOLD}${BLUE}│${RESET}"
p "${BOLD}${BLUE}╰──────────────────────────────────────────────────────╯${RESET}"
p ""
s 0.8

# ═══════════════════════════════════════════════════════════════════════════════
# SCENE 4 — watch mode
# ═══════════════════════════════════════════════════════════════════════════════

slow 0.02 "${BOLD}${CYAN}\$ ci-debugger run --watch -W testdata/demo_v04.yml${RESET}"
s 0.5

p ""
p "${BOLD}CI Pipeline${RESET}  ${DIM}push${RESET}"
p ""
s 0.2
p "${BOLD}▶ build${RESET}"
s 0.2
p "  ${GREEN}✓ [1/4]${RESET} Checkout"
p "  ${GREEN}✓ [2/4]${RESET} Show context"
p "  ${GREEN}✓ [3/4]${RESET} Set build metadata"
p "  ${GREEN}✓ [4/4]${RESET} Build"
p ""
p "${BOLD}▶ test (go=1.22)${RESET}  ${BOLD}▶ test (go=1.23)${RESET}"
s 0.2
p "  ${GREEN}✓${RESET} all steps passed        ${GREEN}✓${RESET} all steps passed"
p ""
p "${BOLD}▶ deploy${RESET}"
s 0.2
p "  ${GREEN}✓ [1/1]${RESET} Deploy 2.1.0"
p ""
p "${DIM}◎ Watching for changes… (Ctrl+C to stop)${RESET}"
s 1.5
p ""
p "${DIM}[testdata/demo_v04.yml changed — re-running]${RESET}"
p ""
s 0.4
p "${BOLD}CI Pipeline${RESET}  ${DIM}push${RESET}"
p ""
s 0.2
p "${BOLD}▶ build${RESET}"
s 0.15
p "  ${GREEN}✓ [1/4]${RESET} Checkout"
p "  ${GREEN}✓ [2/4]${RESET} Show context"
p "  ${GREEN}✓ [3/4]${RESET} Set build metadata"
p "  ${GREEN}✓ [4/4]${RESET} Build"
p ""
p "  ${DIM}Job outputs:${RESET}  version → ${CYAN}2.1.0${RESET}   artifact → ${CYAN}myapp-2.1.0.tar.gz${RESET}"
p ""
p "${DIM}◎ Watching for changes… (Ctrl+C to stop)${RESET}"
s 1.2
p ""
