package healthchecks

import "context"

type Healthcheck interface {
	Healthy() bool
	Init()                 // set flags
	Start(context.Context) // start worker (after parsing flags)
	Update() bool          // set Healthy from worker state
}
