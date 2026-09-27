package request

import "encoding/json"

type PasskeyVerification struct {
	CeremonyToken string          `json:"ceremonyToken"`
	Credential    json.RawMessage `json:"credential" swaggertype:"object"`
	Name          string          `json:"name,omitempty"`
}
