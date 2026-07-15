package api

type apiGroup struct {
	LoginBackground loginBackgroundAPI
}

var Api = new(apiGroup)
