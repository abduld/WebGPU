package jobs

import (
	"github.com/revel/revel"
	"github.com/revel/revel/modules/jobs/app/jobs"
)

func InitJobs() {

	revel.OnAppStart(func() {
		jobs.Schedule("@every 1m", RegisterWorkerJob{})
	})
}
