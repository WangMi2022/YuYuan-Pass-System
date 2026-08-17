package service

import (
	"github.com/WangMi2022/mit-assets-admin/server/service/example"
	"github.com/WangMi2022/mit-assets-admin/server/service/system"
)

var ServiceGroupApp = new(ServiceGroup)

type ServiceGroup struct {
	SystemServiceGroup  system.ServiceGroup
	ExampleServiceGroup example.ServiceGroup
}
