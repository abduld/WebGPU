package bigcode

import (
	"sync"
	. "wb/app/config"
)

func InitBigCode() {
	if BigCodeEnabled {
		databaseCache = map[int]BigCodeDatabase{}
		databaseCacheMutex = &sync.Mutex{}
		initBoldBackend()
	}
}
