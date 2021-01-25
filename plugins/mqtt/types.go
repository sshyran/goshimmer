package mqtt

// message defines the message topic
type message struct {
	// The hex encoded message ID of the message.
	MessageID string `json:"messageId"`
	// The issuer ID of the message.
	IssuerID string `json:"issuerId"`
	// The timestamp of issuance.
	Timestamp int64 `json:"timestamp"`
	// The message IDs of the strong parents the message references.
	StrongParents []string `json:"strongParentsIDs"`
	// The message IDs of the weak parents the message references.
	WeakParents []string `json:"weakParentsIDs"`
	// The Payload of the message.
	Payload []byte `json:"payload"`
	// The nonce of the message.
	Nonce uint64 `json:"nonce"`
	// The signature of the message.
	Signature string `json:"signature"`
}
