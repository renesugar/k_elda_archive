package wait

import (
	"time"

	"github.com/quilt/quilt/util"
)

// Wait provides reasonable default values for `util.BackoffWaitFor` for use
// by provider implementations.
func Wait(pred func() bool) error {
	return util.BackoffWaitFor(pred, 30*time.Second, 5*time.Minute)
}
