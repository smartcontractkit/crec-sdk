package transact

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/cvn-sdk/client"
	"github.com/smartcontractkit/cvn-sdk/internal/mockserver"
	"github.com/smartcontractkit/cvn-sdk/transact/signer"
	"github.com/smartcontractkit/cvn-sdk/transact/types"
)

func TestHashOperation(t *testing.T) {
	// changing these will change the expected hash at the end of this test
	chainId := uint64(31337)
	to := common.HexToAddress("0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f")
	account := common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3")

	mockServer := mockserver.NewMockServer(t)
	defer mockServer.Close()

	c, err := client.NewClientWithResponses(mockServer.TestServer.URL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	transact, err := NewClient(
		c,
		&ClientOptions{
			ChainID: chainId,
		},
	)
	require.NoError(t, err)

	operation := &types.Operation{
		ID:      big.NewInt(1),
		Account: &account,
		Transactions: []*types.Transaction{
			{
				To:    &to,
				Value: big.NewInt(0),
				Data:  []byte(""),
			},
		},
	}

	hash, err := transact.HashOperation(operation)
	if err != nil {
		t.Fatalf("Failed to hash operation: %v", err)
	}

	// check for pre-computed hash for the operation based on the above to/account
	require.Equal(t, "cd4308149652087bf9621b30e3d7781c475abb327b12b4e257966e88fa4a1ada", common.Bytes2Hex(hash))
}

func TestSignOperation(t *testing.T) {
	// changing these will change the expected hash at the end of this test
	chainId := uint64(31337)
	to := common.HexToAddress("0x23618e81E3f5cdF7f54C3d65f7FBc0aBf5B21E8f")
	account := common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3")

	c, err := client.NewClientWithResponses("http://localhost:8080")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	transact, err := NewClient(
		c,
		&ClientOptions{
			ChainID: chainId,
		},
	)
	require.NoError(t, err)

	operation := &types.Operation{
		ID:      big.NewInt(1),
		Account: &account,
		Transactions: []*types.Transaction{
			{
				To:    &to,
				Value: big.NewInt(0),
				Data:  []byte(""),
			},
		},
	}

	privateKey, err := crypto.HexToECDSA("165fdaa699776c9bfdc194817c479d0775b1ee9718bfcddb0ccca352ece86066")
	require.NoError(t, err)

	localSigner := signer.NewLocalSigner(privateKey)
	sig, err := transact.SignOperation(operation, localSigner)
	require.NoError(t, err)

	// check for pre-computed signature for the operation based on the above to/account and private key
	require.Equal(
		t,
		"5e1d5b835e963051f75e33bb8d20dd6464afe89268d53cfc06f3223ffcc1357b30f5fe9f75ceddf99792d9e1c877a3824bef0f79d522985723df46f3185ec75f1b",
		common.Bytes2Hex(sig),
	)
}
