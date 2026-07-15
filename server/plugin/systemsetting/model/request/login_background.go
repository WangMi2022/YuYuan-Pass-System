package request

type CreateBackground struct {
	Name string `json:"name" binding:"required"`
	URL  string `json:"url" binding:"required"`
}

type ActivateBackground struct {
	ID uint `json:"id" binding:"required"`
}
