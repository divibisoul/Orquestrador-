# Self-hosted runner

The recovery branch is configured to execute CI, security scans and benchmarks on a self-hosted GitHub Actions runner.

## Expected runner profile

- GitHub Actions runner: 2.336.0 or newer.
- Linux x64.
- Go 1.23 toolchain available or installable by `actions/setup-go`.
- Python 3 available for the experimental adapter checks.
- Docker available for the Gosec container action.

## Registration

Use a fresh, short-lived registration token from GitHub. Never commit or paste a live token into the repository.

```bash
mkdir actions-runner && cd actions-runner
curl -o actions-runner-linux-x64-2.336.0.tar.gz -L https://github.com/actions/runner/releases/download/v2.336.0/actions-runner-linux-x64-2.336.0.tar.gz
echo "<SHA256>  actions-runner-linux-x64-2.336.0.tar.gz" | sha256sum -c -
tar xzf ./actions-runner-linux-x64-2.336.0.tar.gz
./config.sh --url https://github.com/divibisoul/Orquestrador- --token <REGISTRATION_TOKEN>
./run.sh
```

The repository workflows use `runs-on: self-hosted`. GitHub also adds default labels such as `self-hosted`, `linux` and `x64` unless default labels were disabled during registration.

## Workload model

The Go and Python jobs may share one runner. With a single runner they are serialized when no idle matching runner is available. Add another registered runner when true parallel execution is required.

## Verification

The recovery CI must prove, on this runner:

```bash
gofmt -l .
go vet ./...
go build ./...
go test ./... -count=1 -race
python3 -m py_compile core/superagi/generator/providers/base.py core/superagi/generator/providers/gemini.py core/neural_fabric/router.py api/gemini.py
```

The security workflow separately runs Gosec over `./...`.
