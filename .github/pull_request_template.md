## Description

<!-- Briefly describe the changes in this PR -->

## Type of Change

- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update
- [ ] Refactoring

## Release Checklist

> **Note:** Only complete this section when this PR will trigger a new SDK release.

### 1. Release candidate (pre-release)
- [ ] Create and push a release-candidate git tag (e.g. `v1.1.1-rc1`) in this repo
- [ ] Wait for SDK CI to pass on that tag (unit tests, static analysis)

### 2. System tests against the RC
- [ ] In `crec-courier-service`, open a PR that bumps the SDK dependency to the RC version (e.g. `v1.1.1-rc1`)
- [ ] Adjust system test code in that PR if needed
- [ ] Ensure all system tests pass (mocked CRE via CTF v2)

### 3. Publish release
- [ ] Create and push the final version tag in this repo (e.g. `v1.1.1`)
- [ ] In `crec-courier-service`, update the SDK dependency to the final version and merge (or open a follow-up PR)
- [ ] Confirm CI is green in `crec-courier-service`