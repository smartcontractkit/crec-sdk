package types

type VerifiableEvent struct {
	Type       string   `json:"type"`
	Payload    string   `json:"verifiable_event"`
	Signatures []string `json:"signatures"`
}
