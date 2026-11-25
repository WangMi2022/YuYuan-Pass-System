package response

type ContactVerificationChannelCapability struct {
	Configured bool   `json:"configured"`
	Enabled    bool   `json:"enabled"`
	Reason     string `json:"reason"`
}

type ContactVerificationCapabilities struct {
	Phone ContactVerificationChannelCapability `json:"phone"`
	Email ContactVerificationChannelCapability `json:"email"`
}
