## Description

<!-- Briefly describe the changes in this PR -->

## Type of Change

- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update
- [ ] Refactoring


## Release Checklist

> **Note:** For changes that require a new SDK release, complete the following steps:

### 1. Create Release Candidate
- [ ] Tag release candidate (e.g., `v1.1.1-rc1`)
- [ ] Verify SDK CI passes (unit tests, static analysis)

### 2. Verify System Tests
- [ ] Open PR in `crec-courier-service` with RC version
- [ ] Update system test logic (if needed)
- [ ] All system tests pass (mocked CRE using CTF v2)

### 3. Finalize Release
- [ ] Publish final SDK tag (e.g., `v1.1.1`)
- [ ] Update `crec-courier-service` to final SDK version
- [ ] Confirm all CI passes in `crec-courier-service`

