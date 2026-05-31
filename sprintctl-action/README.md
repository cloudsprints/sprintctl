# CloudSprints Grader Action

Grade CloudSprints labs automatically via pull requests. No student tokens needed — the action identifies students by their Git commit email.

## How It Works

1. Name your branch `grade/<lesson-token>` (copy from CloudSprints UI)
2. Push your code and open a PR
3. The action detects your email from the Git commit, fetches the lesson, runs validation commands, and submits for AI grading
4. Results appear in the workflow log (and optionally as PR comments via the CloudSprints GitHub App)

**Prerequisites:** Your Git email must match the email you signed up with on CloudSprints.

## Quick Start

### 1. Add the workflow file to your repo

Create `.github/workflows/grade.yml`:

```yaml
name: CloudSprints Grader

on:
  pull_request:
    branches: [main]

jobs:
  grade:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: morethancertified/sprintctl-action@v1
        with:
          api-key: ${{ secrets.CLOUDSPRINTS_API_KEY }}
```

### 2. Add the API key (instructor/org admin sets this once)

The `CLOUDSPRINTS_API_KEY` is a shared key that authenticates the GitHub Action to the CloudSprints API. It is **not** a per-student token.

1. Get the API key from your CloudSprints admin
2. In your GitHub repo → Settings → Secrets → Actions → New repository secret
3. Name: `CLOUDSPRINTS_API_KEY`, Value: the key

### 3. Create a grading branch

Copy the branch name from the lesson in CloudSprints:

```bash
git checkout -b grade/cm4ppz694200blze51ts1234
```

### 4. Do your work, push, and open a PR

```bash
# ... complete the lab tasks ...
git add .
git commit -m "completed lab"
git push origin grade/cm4ppz694200blze51ts1234
# Open a PR to main
```

The grading action runs automatically. Check the Actions tab for results.

## Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `api-key` | ✅ | — | CloudSprints API key (shared, set by instructor) |
| `api-url` | ❌ | `https://cloudsprints.com` | API base URL |
| `branch-prefix` | ❌ | `grade/` | Branch prefix that triggers grading |
| `callback-url` | ❌ | — | URL for GitHub App result reporting |

## Outputs

| Output | Description |
|--------|-------------|
| `lesson-token` | Lesson token extracted from branch |
| `email` | Student email detected from Git commit |
| `passed` | Number of tasks passed |
| `failed` | Number of tasks failed |
| `total` | Total number of tasks |
| `status` | `pass` or `fail` |
| `results-json` | Full grading results as JSON |

## Multiple Labs in One Repo

Each lab gets its own branch. The workflow file stays the same:

```
main
├── grade/cm4ppz694200blze51ts1234   ← Lab 1
├── grade/cm5qqz694200blze51ts5678   ← Lab 2
└── grade/cm6rrz694200blze51ts9012   ← Lab 3
```

## With Starter Code

If a lab has starter code, download it into your branch:

```bash
git checkout -b grade/cm4ppz694200blze51ts1234
curl -sL https://cloudsprints.com/labs/k8s-pods-101/starter.tar.gz | tar xz --strip-components=1
git add .
git commit -m "add starter code"
```

## Advanced: Using Outputs

```yaml
- uses: morethancertified/sprintctl-action@v1
  id: grading
  with:
    api-key: ${{ secrets.CLOUDSPRINTS_API_KEY }}

- name: Custom summary
  if: always()
  run: |
    echo "Score: ${{ steps.grading.outputs.passed }}/${{ steps.grading.outputs.total }}"
    echo "Status: ${{ steps.grading.outputs.status }}"
```
