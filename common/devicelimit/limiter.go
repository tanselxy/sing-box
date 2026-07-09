package devicelimit

import (
	"fmt"
	"sync"

	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

var global = &Limiter{
	users: map[string]map[string]int{},
}

type Limiter struct {
	mu    sync.Mutex
	users map[string]map[string]int
}

func Acquire(user string, limit int, source M.Socksaddr) (func(), error) {
	return global.Acquire(user, limit, source)
}

func ReleaseOnClose(onClose N.CloseHandlerFunc, release func()) N.CloseHandlerFunc {
	if release == nil {
		return onClose
	}
	return N.AppendClose(onClose, func(error) {
		release()
	})
}

func (l *Limiter) Acquire(user string, limit int, source M.Socksaddr) (func(), error) {
	if limit <= 0 || user == "" {
		return nil, nil
	}
	device := source.AddrString()
	if device == "" {
		device = fmt.Sprint(source)
	}
	if device == "" {
		device = "unknown"
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	devices := l.users[user]
	if devices == nil {
		devices = map[string]int{}
		l.users[user] = devices
	}
	if count := devices[device]; count > 0 {
		devices[device] = count + 1
		return l.release(user, device), nil
	}
	if len(devices) >= limit {
		return nil, fmt.Errorf("device limit exceeded for user %s: %d/%d active devices", user, len(devices), limit)
	}
	devices[device] = 1
	return l.release(user, device), nil
}

func (l *Limiter) release(user string, device string) func() {
	var once sync.Once
	return func() {
		once.Do(func() {
			l.mu.Lock()
			defer l.mu.Unlock()
			devices := l.users[user]
			if devices == nil {
				return
			}
			count := devices[device]
			if count <= 1 {
				delete(devices, device)
				if len(devices) == 0 {
					delete(l.users, user)
				}
				return
			}
			devices[device] = count - 1
		})
	}
}
