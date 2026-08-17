package request

type CreateBackground struct {
	Name string `json:"name" binding:"required"`
	URL  string `json:"url" binding:"required"`
}

type ActivateBackground struct {
	ID uint `json:"id" binding:"required"`
}

type SaveLoginLogo struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	SystemName string `json:"systemName"`
	Subtitle   string `json:"subtitle"`
}
