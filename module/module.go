package module

import (
	"sync"
)

type Module struct {
	wg sync.WaitGroup
}
