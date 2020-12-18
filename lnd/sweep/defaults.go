// +build !rpctest

package sweep

import (
	"time"
)

// DefaultBatchWindowDuration specifies duration of the sweep batch
// window. The sweep is held back during the batch window to allow more
// inputs to be added and thereby lower the fee per input.
var DefaultBatchWindowDuration = 30 * time.Second
