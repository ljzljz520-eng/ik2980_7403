# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
?   	emergencycomms/cmd/relay	[no test files]
?   	emergencycomms/internal/report	[no test files]
ok  	emergencycomms/internal/codec	0.001s
ok  	emergencycomms/internal/model	0.001s
ok  	emergencycomms/internal/security	0.001s
--- FAIL: TestRepeatedNonceMessageIsRejected (0.00s)
    workflow_test.go:149: repeated nonce was accepted: {Accepted:true Outcome:accepted Reason:authenticated Entry:{ID:765d790f76572cbd-002 Channel:ALPHA-01 Batch:BATCH-1 Nonce:N-1 Outcome:accepted Reason:authenticated Sequence:2 Fingerprint:dfa3a3431182d42eba1882ca676c2251e88c99a80cb493ae5efbd578af293291}}
FAIL
FAIL	emergencycomms/internal/service	0.039s
ok  	emergencycomms/internal/store	0.023s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/relay): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/relay): exit `0`
