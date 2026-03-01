package tools

import (
	"fmt"
	"sync/atomic"
	"time"
)

var toolCallSeq uint64

func nextToolCallID() string {
	n := atomic.AddUint64(&toolCallSeq, 1)
	return fmt.Sprintf("tc-%d-%d", time.Now().UnixNano(), n)
}
