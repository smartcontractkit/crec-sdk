package types

type VerifiableEvent struct {
	Type       string   `json:"type"`
	Payload    string   `json:"verifiableEvent"`
	Signatures []string `json:"signatures"`
}
