package api

var Api = new(apiGroup)

type apiGroup struct {
	WorkSchedule workScheduleAPI
}
