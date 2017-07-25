package audit

import "gitlab.booking.com/infra/dora/simpleapi"

type Audit struct {
	username  string
	password  string
	simpleAPI *simpleapi.SimpleAPI
}

func New(username string, password string, simpleApi *simpleapi.SimpleAPI) *ChassisConnection {
	return &Audit{username: username, password: password, simpleApi: simpleapi}
}
