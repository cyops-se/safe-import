// +build darwin

package usvc

import (
	"time"
	"github.com/cyops-se/safe-import/usvc/types"
	"github.com/nats-io/nats.go"
)

type UsvcBroker struct {
	dependencies  []string
	hostname      string
	services      map[string]Usvc
	servicestates map[string]*types.UsvcState
	lastcheck	  time.Time
	connection    *nats.Conn
	timeout       uint // seconds
	err           error
}
