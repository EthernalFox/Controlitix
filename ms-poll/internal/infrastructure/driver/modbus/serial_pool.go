package modbus

import "sync"

type SerialPortPool struct {
	locks map[string]*sync.Mutex
	mu    sync.Mutex
}

func NewSerialPortPool() *SerialPortPool {
	return &SerialPortPool{
		locks: make(map[string]*sync.Mutex),
	}
}

func (pool *SerialPortPool) Acquire(port string) *sync.Mutex {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	lock, exists := pool.locks[port]
	if exists {
		return lock
	}

	lock = &sync.Mutex{}
	pool.locks[port] = lock
	return lock
}
