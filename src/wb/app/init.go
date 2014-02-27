package app

import (
	"runtime"
	"wb/app/config"
	"wb/app/controllers"
	"wb/app/models"
	"wb/app/server"
	"wb/app/stats"

	"github.com/robfig/revel"
)

func init() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	// Filters is the default set of global filters.
	revel.Filters = []revel.Filter{
		revel.PanicFilter,             // Recover from panics and display an error page instead.
		revel.RouterFilter,            // Use the routing table to select the right Action
		revel.FilterConfiguringFilter, // A hook for adding or removing per-Action filters.
		revel.ParamsFilter,            // Parse parameters into Controller.Params.
		revel.SessionFilter,           // Restore and write the session cookie.
		revel.FlashFilter,             // Restore and write the flash cookie.
		revel.ValidationFilter,        // Restore kept validation errors and save new ones from cookie.
		//revel.I18nFilter,              // Resolve the requested language
		revel.InterceptorFilter, // Run interceptors around the action.
		revel.ActionInvoker,     // Invoke the action.
	}

	revel.OnAppStart(func() {
		config.InitConfig()
		stats.InitStats()
		controllers.InitControllers()
		models.InitModels()
		server.InitServer()

		stats.Incr("Server", "Startup")

	})
}
