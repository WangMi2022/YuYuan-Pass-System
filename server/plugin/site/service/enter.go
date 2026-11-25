package service

var Service = new(serviceGroup)

type serviceGroup struct {
	Site siteService
}
