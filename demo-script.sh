#!/usr/bin/env bash
# Simulates ci-debugger run --step output for demo GIF

RESET='\033[0m'
BOLD='\033[1m'
DIM='\033[90m'
GREEN='\033[32m'
RED='\033[31m'
YELLOW='\033[33m'
PURPLE='\033[35m'
CYAN='\033[36m'

sleep_ms() { sleep "0.$1"; }

echo ""
echo -e "${BOLD}ci-debugger${RESET}  My CI Workflow  ${DIM}on: push${RESET}"
echo -e "${DIM}───────────────────────────────────────────────────────${RESET}"
sleep_ms 400

echo ""
echo -e "${PURPLE}${BOLD}▶ test${RESET}  ${DIM}(ghcr.io/catthehacker/ubuntu:act-22.04)${RESET}"
sleep_ms 600

echo -e "  ${YELLOW}⟳${RESET} ${DIM}[1/4]${RESET} Checkout"
sleep_ms 300
echo -e "  ${GREEN}✓${RESET} ${DIM}[1/4]${RESET} Checkout  ${DIM}(0.2s)${RESET}"
sleep_ms 500

echo -e "  ${YELLOW}⟳${RESET} ${DIM}[2/4]${RESET} Install dependencies"
sleep_ms 800
echo -e "  ${GREEN}✓${RESET} ${DIM}[2/4]${RESET} Install dependencies  ${DIM}(8.3s)${RESET}"
sleep_ms 500

# Breakpoint before "Run tests"
echo ""
echo -e "${YELLOW}${BOLD}◆ BREAKPOINT${RESET}  before step ${BOLD}Run tests${RESET}"
echo -e "  ${DIM}Command:${RESET}"
echo -e "  ${DIM}  pytest -x --tb=short -q${RESET}"
echo ""
echo -e "  ${BOLD}[C]${RESET} Continue  ${BOLD}[S]${RESET} Skip  ${BOLD}[D]${RESET} Shell  ${BOLD}[I]${RESET} Inspect  ${BOLD}[Q]${RESET} Quit"
echo -n "  → "
sleep 1.5

# User types "d" to open shell
echo -e "d"
sleep_ms 500

echo ""
echo -e "  ${YELLOW}[ci-debugger] Dropped into container shell. Type 'exit' to return.${RESET}"
echo ""
echo -e "${CYAN}root@a1b2c3:/github/workspace#${RESET} ls"
sleep_ms 600
echo -e "src/  tests/  requirements.txt  pytest.ini"
sleep_ms 400
echo -e "${CYAN}root@a1b2c3:/github/workspace#${RESET} cat pytest.ini"
sleep_ms 500
echo -e "[pytest]"
echo -e "testpaths = tests"
sleep_ms 400
echo -e "${CYAN}root@a1b2c3:/github/workspace#${RESET} exit"
sleep_ms 600
echo ""

# Re-show prompt after shell
echo -e "  ${BOLD}[C]${RESET} Continue  ${BOLD}[S]${RESET} Skip  ${BOLD}[D]${RESET} Shell  ${BOLD}[I]${RESET} Inspect  ${BOLD}[Q]${RESET} Quit"
echo -n "  → "
sleep 1.2

echo -e "c"
sleep_ms 600

echo -e "  ${YELLOW}⟳${RESET} ${DIM}[3/4]${RESET} Run tests"
sleep_ms 400
echo -e "  ${GREEN}✓${RESET} ${DIM}[3/4]${RESET} Run tests  ${DIM}(12.1s)${RESET}"
sleep_ms 400

echo -e "  ${YELLOW}⟳${RESET} ${DIM}[4/4]${RESET} Build"
sleep_ms 500
echo -e "  ${GREEN}✓${RESET} ${DIM}[4/4]${RESET} Build  ${DIM}(3.4s)${RESET}"
sleep_ms 600

# Summary
echo ""
echo -e "${DIM}───────────────────────────────────────────────────────${RESET}"
echo -e "${BOLD}Summary${RESET}"
echo ""
echo -e "╭──────────────────────────────────────────────────────────╮"
echo -e "│                                                          │"
echo -e "│  ${PURPLE}Job: test${RESET}  ${GREEN}(passed)${RESET}                                    │"
echo -e "│  1   Checkout              ${GREEN}passed${RESET}      0.2s           │"
echo -e "│  2   Install dependencies  ${GREEN}passed${RESET}      8.3s           │"
echo -e "│  3   Run tests             ${GREEN}passed${RESET}     12.1s           │"
echo -e "│  4   Build                 ${GREEN}passed${RESET}      3.4s           │"
echo -e "│                                                          │"
echo -e "╰──────────────────────────────────────────────────────────╯"
echo ""
echo -e "  Total: ${BOLD}24.0s${RESET}  ${GREEN}4 passed${RESET}"
echo ""
