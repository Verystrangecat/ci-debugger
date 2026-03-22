# ci-debugger

**Debug GitHub Actions workflows locally — with breakpoints.**

You shouldn't have to push 47 commits to figure out why your CI is failing.

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
| Clean, readable output | ✓ | ✗ (70K log lines) |
| GITHUB_OUTPUT support | ✓ | ✓ |
| `actions/checkout` support | ✓ (workspace mounted) | ✓ |
| Full actions/\* support | Partial (coming soon) | Partial |
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

# List available workflows
ci-debugger list

# Run with step-by-step mode
ci-debugger run --step

# Run specific workflow
ci-debugger run -W .github/workflows/ci.yml

# Run specific job
ci-debugger run -j test

# Break when anything fails
ci-debugger run --break-on-error

# Break before a specific step
ci-debugger run --break-before "Run tests"

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

- **Linux runners only** — `windows-latest` and `macos-latest` are not supported
- **`uses:` actions** — `actions/checkout` is handled automatically (workspace is mounted). Other actions are skipped with a warning. Convert to `run:` steps as a workaround
- **Expression evaluation** — Basic `${{ env.X }}` and `${{ secrets.X }}` are supported. Complex expressions may not evaluate correctly

These are all planned improvements. Contributions welcome!

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
