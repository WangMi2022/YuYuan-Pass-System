package router

import "github.com/WangMi2022/mit-assets-admin/server/plugin/smart/api"

var Router = new(routerGroup)

type routerGroup struct{ Smart smartRouter }

var apiSmart = api.SmartAPI
