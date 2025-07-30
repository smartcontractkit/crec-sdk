#!/bin/bash

echo "Running Vault Transit Signer Tests..."
echo "======================================"

echo ""
echo "1. Running standalone Transit Signer unit tests..."
go test ./transact/signer -v -run TestTransitSigner -timeout=60s

echo ""
echo "2. Running integration test with transact client..."
go test ./transact -v -run TestSignOperationWithVaultTransit -timeout=60s

echo ""
echo "All Vault Transit tests completed!"
echo ""
echo "Note: These tests use testcontainers to spin up real Vault instances"
echo "and test the Transit secrets engine with RSA-2048 keys."