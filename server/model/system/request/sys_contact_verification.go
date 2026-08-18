package request

type SendContactVerificationCode struct {
	Channel string `json:"channel"`
	Target  string `json:"target"`
}

type UpdateSelfContact struct {
	Channel string `json:"channel"`
	Target  string `json:"target"`
	Code    string `json:"code"`
}
