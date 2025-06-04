package signer

type Signer interface {
	Sign(hash []byte) ([]byte, error)
}
