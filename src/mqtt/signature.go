package mqtt

// Signature contains the signature data
type Signature struct {
	Signature string        `json:"signature"`
	Meta      SignatureMeta `json:"meta"`
}

// SignatureMeta contains the meta information from the signature
type SignatureMeta struct {
	Date      string `json:"date"`
	Algorithm string `json:"algorithm"`
}
