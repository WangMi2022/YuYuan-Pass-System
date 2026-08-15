package router

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/smart/api"

var Router = new(routerGroup)

type routerGroup struct{ Smart smartRouter }

var apiSmart = api.SmartAPI
