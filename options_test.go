package crec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithDONConfig_AppliesAtomicUnit(t *testing.T) {
	signers := []string{"0x5db070ceabcf97e45d96b4f951a1df050ddb5559", "0xadebb9657c04692275973230b06adfabacc899bc"}

	cfg := &clientConfig{}
	WithDONConfig("3", 2, signers)(cfg)

	assert.Equal(t, "3", cfg.creTenantID)
	assert.Equal(t, 2, cfg.minRequiredSignatures)
	assert.Equal(t, signers, cfg.validSigners)
	assert.True(t, cfg.donConfigSet)
}
