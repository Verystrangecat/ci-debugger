#!/bin/bash
# ci-debugger full feature demo — simulated output for GIF recording

G='\033[32m'   # green
R='\033[31m'   # red
Y='\033[33m'   # yellow
C='\033[36m'   # cyan
D='\033[90m'   # dim/gray
B='\033[1m'    # bold
M='\033[35m'   # magenta
Z='\033[0m'    # reset

# ─────────────────────────────────────────────────────────────
# Section 1: list
# ─────────────────────────────────────────────────────────────
echo -e "\n${B}$ ci-debugger list -W testdata/simple.yml${Z}"
sleep 0.4

echo -e "\n${D}Workflows in testdata/${Z}\n"
echo -e "  simple.yml  ${B}Simple Test${Z}"
echo -e "    └─ ${C}test${Z}  ${D}(ubuntu-latest)${Z}"
echo -e "       ├─ [1] Say hello"
echo -e "       ├─ [2] Show environment"
echo -e "       └─ [3] Multi-line script"
sleep 2.5

# ─────────────────────────────────────────────────────────────
# Section 2: scan
# ─────────────────────────────────────────────────────────────
echo -e "\n${B}$ ci-debugger scan${Z}"
sleep 0.4

echo -e "\n${B}Scan Results${Z}  ${D}(3 workflow(s))${Z}"
echo -e "${D}────────────────────────────────────────────────────────────────${Z}"
echo -e "  ${G}✓ No issues found${Z}"
echo -e "${D}────────────────────────────────────────────────────────────────${Z}"
echo -e "${G}  0 error(s), 0 warning(s)${Z}\n"
sleep 2.5

# ─────────────────────────────────────────────────────────────
# Section 3: --env-report
# ─────────────────────────────────────────────────────────────
echo -e "${B}$ ci-debugger run --env-report -W testdata/simple.yml${Z}"
sleep 0.4

echo -e "\n${B}Environment Variables Report${Z}"
echo -e "${D}────────────────────────────────────────────────────────────────────────${Z}"
printf "  ${B}%-38s %-14s %s${Z}\n" "Variable" "Status" "Value"
echo -e "${D}────────────────────────────────────────────────────────────────────────${Z}"
printf "  %-38s ${G}%-14s${Z} %s\n" "GITHUB_ACTIONS"              "real"        "true"
printf "  %-38s ${G}%-14s${Z} %s\n" "GITHUB_SHA"                  "real"        "a3f4b2c1d..."
printf "  %-38s ${G}%-14s${Z} %s\n" "GITHUB_REPOSITORY"           "real"        "owner/repo"
printf "  %-38s ${G}%-14s${Z} %s\n" "GITHUB_REF"                  "real"        "refs/heads/main"
printf "  %-38s ${Y}%-14s${Z} %s\n" "GITHUB_WORKSPACE"            "stubbed"     "/github/workspace"
printf "  %-38s ${Y}%-14s${Z} %s\n" "RUNNER_OS"                   "stubbed"     "Linux"
printf "  %-38s ${Y}%-14s${Z} %s\n" "RUNNER_ARCH"                 "stubbed"     "X64"
printf "  %-38s ${R}%-14s${Z} %s\n" "GITHUB_TOKEN"                "unavailable" "(injected by GitHub)"
printf "  %-38s ${R}%-14s${Z} %s\n" "ACTIONS_ID_TOKEN_REQUEST_URL" "unavailable" "(OIDC not available locally)"
echo -e "${D}────────────────────────────────────────────────────────────────────────${Z}"
echo -e "\n  ${D}Legend:  ${G}real${Z}${D} = from local git   ${Y}stubbed${Z}${D} = fixed local value   ${R}unavailable${Z}${D} = GitHub-only${Z}\n"
sleep 4

# ─────────────────────────────────────────────────────────────
# Section 4: matrix builds
# ─────────────────────────────────────────────────────────────
echo -e "${B}$ ci-debugger run -W testdata/matrix.yml${Z}"
sleep 0.4

echo -e "\n${M}${B}ci-debugger${Z}  ${B}Matrix CI${Z}  ${D}on: push${Z}"
echo -e "${D}───────────────────────────────────────────────────────${Z}\n"
echo -e "  Matrix: 2 combination(s) for job ${B}\"test\"${Z}\n"
sleep 0.4

# Job 1: node-version=18
echo -e "\n${B}${C}▶ test (node-version=18, os=ubuntu-latest)${Z}  ${D}(ghcr.io/catthehacker/ubuntu:act-latest)${Z}"
sleep 0.2
echo -e "  ${Y}⟳${Z} ${D}[1/3]${Z} Checkout"
sleep 0.4
echo -e "  ${G}✓${Z} ${D}[1/3]${Z} Checkout  ${D}(98ms)${Z}"
echo -e "  ${Y}⟳${Z} ${D}[2/3]${Z} Use Node.js 18"
sleep 0.5
echo -e "  ${G}✓${Z} ${D}[2/3]${Z} Use Node.js 18  ${D}(210ms)${Z}"
echo -e "  ${Y}⟳${Z} ${D}[3/3]${Z} Run tests"
sleep 1.2
echo -e "  ${G}✓${Z} ${D}[3/3]${Z} Run tests  ${D}(1.34s)${Z}"
echo -e "  ${G}✓ Job passed${Z}  ${B}test${Z}  ${D}(1.95s)${Z}"
sleep 0.5

# Job 2: node-version=20
echo -e "\n${B}${C}▶ test (node-version=20, os=ubuntu-latest)${Z}  ${D}(ghcr.io/catthehacker/ubuntu:act-latest)${Z}"
sleep 0.2
echo -e "  ${Y}⟳${Z} ${D}[1/3]${Z} Checkout"
sleep 0.3
echo -e "  ${G}✓${Z} ${D}[1/3]${Z} Checkout  ${D}(91ms)${Z}"
echo -e "  ${Y}⟳${Z} ${D}[2/3]${Z} Use Node.js 20"
sleep 0.4
echo -e "  ${G}✓${Z} ${D}[2/3]${Z} Use Node.js 20  ${D}(188ms)${Z}"
echo -e "  ${Y}⟳${Z} ${D}[3/3]${Z} Run tests"
sleep 1.0
echo -e "  ${G}✓${Z} ${D}[3/3]${Z} Run tests  ${D}(1.12s)${Z}"
echo -e "  ${G}✓ Job passed${Z}  ${B}test${Z}  ${D}(1.73s)${Z}"

echo -e "\n${D}───────────────────────────────────────────────────────${Z}"
echo -e "  Total: ${B}3.68s${Z}  ${G}6 passed${Z}"
sleep 3

# ─────────────────────────────────────────────────────────────
# Section 5: Azure DevOps pipeline
# ─────────────────────────────────────────────────────────────
echo -e "\n${B}$ ci-debugger run -W testdata/azure-pipelines.yml${Z}"
sleep 0.4

echo -e "${D}Detected Azure DevOps pipeline: testdata/azure-pipelines.yml${Z}\n"
echo -e "${M}${B}ci-debugger${Z}  ${B}azure-pipelines.yml${Z}"
echo -e "${D}───────────────────────────────────────────────────────${Z}"

echo -e "\n${B}${C}▶ Build${Z}  ${D}(ghcr.io/catthehacker/ubuntu:act-latest)${Z}"
sleep 0.2
echo -e "  ${Y}⟳${Z} ${D}[1/3]${Z} Checkout"
sleep 0.3
echo -e "  ${G}✓${Z} ${D}[1/3]${Z} Checkout  ${D}(87ms)${Z}"
echo -e "  ${Y}⟳${Z} ${D}[2/3]${Z} Build"
sleep 0.9
echo -e "  ${G}✓${Z} ${D}[2/3]${Z} Build  ${D}(912ms)${Z}"
echo -e "  ${Y}⟳${Z} ${D}[3/3]${Z} Run unit tests"
sleep 1.0
echo -e "  ${G}✓${Z} ${D}[3/3]${Z} Run unit tests  ${D}(1.04s)${Z}"
echo -e "  ${G}✓ Job passed${Z}  ${B}Build${Z}  ${D}(2.04s)${Z}"

echo -e "\n${D}───────────────────────────────────────────────────────${Z}"
echo -e "  Total: ${B}2.04s${Z}  ${G}3 passed${Z}"
sleep 3

# ─────────────────────────────────────────────────────────────
# Section 6: step debugger + interactive shell
# ─────────────────────────────────────────────────────────────
echo -e "\n${B}$ ci-debugger run --step -W testdata/failing.yml${Z}"
sleep 0.4

echo -e "\n${Y}${B}◆ Debugger enabled${Z}  ${D}(step mode)${Z}\n"
echo -e "${M}${B}ci-debugger${Z}  ${B}Failing Workflow${Z}  ${D}on: push${Z}"
echo -e "${D}───────────────────────────────────────────────────────${Z}"
echo -e "\n${B}${C}▶ fail-test${Z}  ${D}(ghcr.io/catthehacker/ubuntu:act-latest)${Z}\n"
sleep 0.5

echo -e "  ${Y}◆${Z} Break before: ${B}[1/3] This passes${Z}\n"
echo -e "    Options: ${B}[C]ontinue${Z}  [S]kip  [D]rop into shell  [I]nspect  [Q]uit"
sleep 1.0
echo -e "    > d\n"
sleep 0.3
echo -e "  ${Y}[ci-debugger] Dropped into container shell. Type 'exit' to return.${Z}\n"
echo -e "  root@c3f1a2b:/github/workspace# ${B}ls${Z}"
sleep 0.5
echo -e "  README.md  go.mod  go.sum  internal/  cmd/"
echo -e "  root@c3f1a2b:/github/workspace# ${B}echo \$GITHUB_SHA${Z}"
sleep 0.4
echo -e "  a3f4b2c1d5e6f789012345678abcdef"
echo -e "  root@c3f1a2b:/github/workspace# ${B}exit${Z}\n"
sleep 0.6

echo -e "  ${G}✓${Z} ${D}[1/3]${Z} This passes  ${D}(231ms)${Z}\n"
sleep 0.5

echo -e "  ${Y}◆${Z} Break before: ${B}[2/3] This fails${Z}\n"
echo -e "    Options: ${B}[C]ontinue${Z}  [S]kip  [D]rop into shell  [I]nspect  [Q]uit"
sleep 0.8
echo -e "    > c\n"
sleep 0.3

echo -e "  ${R}✗${Z} ${D}[2/3]${Z} This fails  ${D}(exit 1, 89ms)${Z}"
echo -e "    ${R}── stderr ──${Z}"
echo -e "  ${D}⊘${Z} ${D}[3/3]${Z} This should be skipped  ${D}(skipped)${Z}"
echo -e "  ${R}✗ Job failed${Z}  ${B}fail-test${Z}  ${D}(1.12s)${Z}\n"
sleep 3
