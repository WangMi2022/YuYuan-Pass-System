package api

type apiGroup struct {
	LoginBackground loginBackgroundAPI
	LoginLogo       loginLogoAPI
}

var Api = new(apiGroup)
