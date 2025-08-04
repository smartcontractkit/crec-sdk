package signer

import "context"

type Signer interface {
	Sign(ctx context.Context, hash []byte) ([]byte, error)
}
