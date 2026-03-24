# ci-debugger

**Debug CI pipelines locally — with breakpoints.**

You shouldn't have to push 47 commits to figure out why your CI is failing.

![ci-debugger demo](assets/demo.gif)

```
ci-debugger run --step
```

---

## The Problem

Every developer knows the loop:

1. Write some YAML
2. Push to GitHub
3. Wait 5 minutes
4. See a cryptic error
5. Repeat

Existing tools like `act` help run workflows locally, but they're missing the one thing that makes debugging *actually useful*: **the ability to pause, inspect, and interact**.

`act` gives you 70,000 lines of unformatted logs and no way to drop into a shell when something goes wrong.

**ci-debugger** fixes that.

---

## Features

### Breakpoints

Pause execution before or after any step — by name:

```bash
ci-debugger run --break-before "Run tests"
ci-debugger run --break-after "Build"
ci-debugger run --break-on-error
```

At each breakpoint you get an interactive prompt:

```
◆ BREAKPOINT  before step Run tests
  Command:
    pytest -x --tb=short

  [C] Continue  [S] Skip  [D] Shell  [I] Inspect  [Q] Quit

  →
```

### Step-by-Step Mode

Execute one step at a time:

```bash
ci-debugger run --step
```

Great for walking through a new workflow for the first time.

### Interactive Shell

Drop into the container at any breakpoint with `[D]`:

```bash
# At any breakpoint, press D
[ci-debugger] Dropped into container shell. Type 'exit' to return.

root@abc123:/github/workspace# ls
src/  tests/  go.mod  go.sum
root@abc123:/github/workspace# echo $GITHUB_SHA
a1b2c3d4...
root@abc123:/github/workspace# exit
```

The container state is preserved — continue from where you left off.

### Matrix Builds

Workflows with `strategy.matrix` are automatically expanded and run as separate jobs, each with its own container:

```yaml
strategy:
  matrix:
    node-version: [18, 20]
    os: [ubuntu-latest]
```

```
Matrix: 2 combination(s) for job "test"

▶ test (node-version=18, os=ubuntu-latest)
  ✓ [1/3] Checkout
  ✓ [2/3] Use Node.js 18
  ✓ [3/3] Run tests

▶ test (node-version=20, os=ubuntu-latest)
  ✓ [1/3] Checkout
  ✓ [2/3] Use Node.js 20
  ✓ [3/3] Run tests
```

Use `${{ matrix.node-version }}` in your steps — expressions are expanded per combination. `fail-fast` is respected (default: true).

### Service Containers

Sidecar services (postgres, redis, mysql, etc.) are started automatically when your job defines them:

```yaml
services:
  postgres:
    image: postgres:15
    env:
      POSTGRES_PASSWORD: secret
```

Services are reachable by hostname (`postgres`, `redis`, etc.) inside your job container via a shared Docker bridge network.

### Composite `uses:` Actions

`actions/checkout` is handled automatically (your workspace is already mounted). For other `uses:` steps, ci-debugger fetches the `action.yml` from GitHub and runs composite actions inline in your container:

```
↓ fetching actions/setup-node@v4...
▶ running composite action (3 step(s))
  ✓ [1/3] Set up Node
  ✓ [2/3] Install dependencies
  ✓ [3/3] Cache
```

### Node & Docker `uses:` Actions

Node (`node20`, `node16`, `node12`) and Docker action types are now executed locally — not just skipped.

**Node actions** are downloaded from GitHub, extracted to `/tmp`, copied into the job container and run with `node`:

```
↓ fetching actions/github-script@v7...
  Running script...
  ✓ actions/github-script@v7  (1.2s)
```

**Docker actions** pull the action's image and run it as a sidecar container with `INPUT_*` env vars and the workspace mounted:

```
↓ fetching docker://alpine:3.18...
  Pulling image alpine:3.18...
  ✓ docker action  (0.8s)
```

Actions that use a `Dockerfile` (build-time) are skipped with a warning — only pre-built images are supported.

### Full `${{ }}` Expression Engine

All expression namespaces are resolved:

| Namespace | Example |
|-----------|---------|
| `env` | `${{ env.DATABASE_URL }}` |
| `secrets` | `${{ secrets.GITHUB_TOKEN }}` |
| `matrix` | `${{ matrix.node-version }}` |
| `inputs` | `${{ inputs.version }}` |
| `github` | `${{ github.sha }}`, `${{ github.ref }}` |
| `needs` | `${{ needs.build.outputs.artifact }}` |
| `steps` | `${{ steps.setup.outputs.path }}` |
| `job` | `${{ job.status }}` |

`if:` conditions support `success()`, `failure()`, `always()`, `cancelled()`, comparisons (`==`, `!=`), logical operators (`&&`, `||`, `!`), and string functions (`contains()`, `startsWith()`, `endsWith()`).

### Job Outputs Propagation

Jobs can declare `outputs:` and downstream jobs read them via `needs.JOB.outputs.KEY`:

```yaml
jobs:
  build:
    outputs:
      version: ${{ steps.tag.outputs.version }}
    steps:
      - id: tag
        run: echo "version=1.2.3" >> $GITHUB_OUTPUT

  deploy:
    needs: build
    steps:
      - run: echo "Deploying ${{ needs.build.outputs.version }}"
```

```
▶ build
  ✓ [1/1] tag

▶ deploy
  Deploying 1.2.3
  ✓ [1/1] run
```

### Watch Mode

Re-run the workflow automatically whenever workflow files or workspace source files change:

```bash
ci-debugger run --watch -W .github/workflows/ci.yml
```

```
◎ Watching for changes… (Ctrl+C to stop)
```

Watches `.go`, `.yml`, `.yaml`, `.ts`, `.js`, `.py`, `.sh`, `.env`, `Makefile`, and `.secrets` files. Skips `vendor/`, `node_modules/`, `.git/`, and `bin/`. Changes are debounced (500ms) to avoid double-triggers on save.

### Azure DevOps Pipelines

Point ci-debugger at an `azure-pipelines.yml` — it auto-detects the format and maps it to the same runner:

```bash
ci-debugger run -W azure-pipelines.yml
```

```
Detected Azure DevOps pipeline: azure-pipelines.yml

▶ Build  (ghcr.io/catthehacker/ubuntu:act-latest)
  ✓ [1/3] Checkout
  ✓ [2/3] Build
  ✓ [3/3] Run unit tests
```

Supports: `script:` / `bash:`, `task:` (skipped with warning), `checkout: self`, `pool.vmImage`, `variables`, `dependsOn`, `condition`, stages, and top-level steps. If there's no `azure-pipelines.yml` in `.github/workflows/`, ci-debugger also looks for it in the project root automatically.

### Static Analysis

Scan your workflows before running them:

```bash
ci-debugger scan
```

```
Scan Results  (3 workflow(s))
────────────────────────────────────────────────────────────────
  ✖ error
    location: ci.yml > deploy > "Deploy to prod"
    needs: references unknown job "release"

  ⚠ warning
    location: ci.yml > test > step 2
    uses: "actions/setup-python@v4" is not supported locally — will be skipped
────────────────────────────────────────────────────────────────
  1 error(s), 1 warning(s)
```

Catches: circular job dependencies, invalid `needs:` references, unsupported `uses:` actions, steps missing `run:` or `uses:`, and unclosed `${{ }}` expressions.

### Env Var Transparency

See exactly which GitHub environment variables are real, stubbed, or unavailable locally before running:

```bash
ci-debugger run --env-report -W .github/workflows/ci.yml
```

```
Environment Variables Report
────────────────────────────────────────────────────────────────────────
  Variable                               Status         Value
────────────────────────────────────────────────────────────────────────
  GITHUB_ACTIONS                         real           true
  GITHUB_SHA                             real           a3f4b2c1d...
  GITHUB_REPOSITORY                      real           owner/repo
  GITHUB_REF                             real           refs/heads/main
  GITHUB_WORKSPACE                       stubbed        /github/workspace
  RUNNER_OS                              stubbed        Linux
  GITHUB_TOKEN                           unavailable    (injected by GitHub)
  ACTIONS_ID_TOKEN_REQUEST_URL           unavailable    (OIDC not available locally)
────────────────────────────────────────────────────────────────────────

  Legend:  real = from local git   stubbed = fixed local value   unavailable = GitHub-only
```

### Beautiful Output

Clean, color-coded output that shows what matters:

```
ci-debugger  My CI Workflow

▶ test  (ghcr.io/catthehacker/ubuntu:act-latest)
  ✓ [1/3] Checkout
  ✓ [2/3] Install dependencies  (12.3s)
  ✗ [3/3] Run tests  (exit 1, 45.2s)
    ── stderr ──
    FAILED tests/test_api.py::test_create_user
    AssertionError: 404 != 200

╭──────────────────────────────────────────────────────╮
│  Summary                                             │
│                                                      │
│  Job: test                                           │
│  1   Checkout              passed    0.1s            │
│  2   Install dependencies  passed   12.3s            │
│  3   Run tests             FAILED   45.2s            │
╰──────────────────────────────────────────────────────╯

  Total: 57.6s  2 passed  1 failed
```

---

## vs act

| Feature | ci-debugger | act |
|---------|-------------|-----|
| Run workflows locally | ✓ | ✓ |
| Breakpoints | ✓ | ✗ |
| Step-by-step mode | ✓ | ✗ |
| Interactive shell at breakpoint | ✓ | ✗ |
| Matrix builds | ✓ | ✓ |
| Service containers | ✓ | ✓ |
| Composite `uses:` actions | ✓ | ✓ |
| Node/Docker `uses:` actions | ✓ | ✓ |
| Full `${{ }}` expression engine | ✓ | ~ |
| Job outputs (`needs.X.outputs.Y`) | ✓ | ✓ |
| Watch mode (`--watch`) | ✓ | ✗ |
| Azure DevOps Pipelines | ✓ | ✗ |
| Static analysis (`scan`) | ✓ | ✗ |
| Env var transparency (`--env-report`) | ✓ | ✗ |
| Clean, readable output | ✓ | ✗ (70K log lines) |
| `GITHUB_OUTPUT` support | ✓ | ✓ |
| Windows runners | ✗ | ✗ |

---

## Installation

### Homebrew (macOS/Linux)

```bash
brew install murataslan1/tap/ci-debugger
```

### Go Install

```bash
go install github.com/murataslan1/ci-debugger@latest
```

### Pre-built Binaries

Download from [Releases](https://github.com/murataslan1/ci-debugger/releases).

### Requirements

- [Docker](https://docker.com) must be running
- Go 1.21+ (for `go install`)

---

## Quick Start

```bash
# In your project directory
cd your-project/

# List available workflows and jobs
ci-debugger list

# Scan for issues before running
ci-debugger scan

# Check which env vars are real vs stubbed
ci-debugger run --env-report

# Run with step-by-step mode
ci-debugger run --step

# Run specific workflow
ci-debugger run -W .github/workflows/ci.yml

# Run an Azure DevOps pipeline
ci-debugger run -W azure-pipelines.yml

# Run specific job
ci-debugger run -j test

# Break when anything fails
ci-debugger run --break-on-error

# Break before a specific step
ci-debugger run --break-before "Run tests"

# Watch mode — re-run on file change
ci-debugger run --watch

# Full output
ci-debugger run -v
```

---

## Configuration

### Environment Variables

Create a `.env` file in your project root:

```bash
DATABASE_URL=postgres://localhost/myapp
API_KEY=dev-key-here
```

### Secrets

Create a `.secrets` file:

```bash
GITHUB_TOKEN=ghp_xxx
NPM_TOKEN=npm_xxx
```

Both files follow `.env` format: `KEY=VALUE` lines, `#` comments, optional `export` prefix.

### Platform Overrides

Map `runs-on` labels to custom Docker images:

```bash
ci-debugger run --platform ubuntu-latest=my-registry/ubuntu:custom
```

### Default Image Mappings

| `runs-on` | Docker Image |
|-----------|-------------|
| `ubuntu-latest` | `ghcr.io/catthehacker/ubuntu:act-latest` |
| `ubuntu-24.04` | `ghcr.io/catthehacker/ubuntu:act-24.04` |
| `ubuntu-22.04` | `ghcr.io/catthehacker/ubuntu:act-22.04` |
| `ubuntu-20.04` | `ghcr.io/catthehacker/ubuntu:act-20.04` |

---

## Known Limitations

- **Linux runners only** — `windows-latest` and `macos-latest` map to the ubuntu image as best-effort
- **Dockerfile-based Docker actions** — actions that build from a `Dockerfile` are skipped with a warning; only pre-built images (`docker://image`) are supported
- **`GITHUB_TOKEN` and OIDC** — not available locally; provide via `.secrets` for workflows that need it

---

## Roadmap

### v0.2 — Core gaps ✓
- [x] [`uses:` composite action support](https://github.com/murataslan1/ci-debugger/issues/1)
- [x] [Env var transparency (`--env-report`)](https://github.com/murataslan1/ci-debugger/issues/3)
- [x] [Static analysis (`ci-debugger scan`)](https://github.com/murataslan1/ci-debugger/issues/4)

### v0.3 — Expansion ✓
- [x] [Azure DevOps Pipelines support](https://github.com/murataslan1/ci-debugger/issues/5)
- [x] [Service containers](https://github.com/murataslan1/ci-debugger/issues/6)
- [x] [Matrix builds](https://github.com/murataslan1/ci-debugger/issues/7)

### v0.4 ✓
- [x] Node/Docker `uses:` action execution
- [x] Full `${{ }}` expression engine
- [x] Job `outputs:` propagation (`needs.JOB.outputs.KEY`)
- [x] `--watch` mode — re-run on file change

Have an idea? [Open an issue](https://github.com/murataslan1/ci-debugger/issues/new).

---

## Contributing

Pull requests are welcome. For major changes, open an issue first.

```bash
git clone https://github.com/murataslan1/ci-debugger
cd ci-debugger
go test ./...
go build -o bin/ci-debugger ./cmd/ci-debugger
```

---

## License

MIT © Murat Aslan
