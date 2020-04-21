package manager

import (
	"sync"

	"github.com/rotisserie/eris"
)

var (
	AsyncManagerExistsError = func(mgr string) error {
		return eris.Errorf("async manager already registered for the given manager %s", mgr)
	}
	InformerAlreadyRegisteredError = func(informer string) error {
		return eris.Errorf("informer already registered with the given name %s", informer)
	}
	InformerNotRegisteredError = eris.New("no informer registered with the given name")
)

/*
	structure to store AsyncManagers in a go routine safe manner
*/
type AsyncManagerMap struct {
	lock     sync.RWMutex
	managers map[string]AsyncManager
}

func NewAsyncManagerMap() *AsyncManagerMap {
	return &AsyncManagerMap{
		managers: make(map[string]AsyncManager),
	}
}

func (c *AsyncManagerMap) SetManager(name string, asyncManager AsyncManager) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if _, ok := c.managers[name]; ok {
		return AsyncManagerExistsError(name)
	}
	c.managers[name] = asyncManager
	return nil
}

func (c *AsyncManagerMap) GetManager(name string) (AsyncManager, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	mgr, ok := c.managers[name]
	return mgr, ok
}

func (c *AsyncManagerMap) ListManagersByName() map[string]AsyncManager {
	c.lock.RLock()
	defer c.lock.RUnlock()
	result := make(map[string]AsyncManager)
	for k, v := range c.managers {
		result[k] = v
	}
	return result
}

func (c *AsyncManagerMap) RemoveManager(name string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.managers, name)
}

/*
	structure to store AsyncManagerHandler in a go routine safe manner
*/
type AsyncManagerHandlerMap struct {
	lock     sync.RWMutex
	handlers map[string]AsyncManagerHandler
}

func NewAsyncManagerHandler() *AsyncManagerHandlerMap {
	return &AsyncManagerHandlerMap{
		handlers: make(map[string]AsyncManagerHandler),
	}
}

func (c *AsyncManagerHandlerMap) SetHandler(name string, mgr AsyncManagerHandler) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if _, ok := c.handlers[name]; ok {
		return InformerAlreadyRegisteredError(name)
	}
	c.handlers[name] = mgr
	return nil
}

func (c *AsyncManagerHandlerMap) GetHandler(name string) (AsyncManagerHandler, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	mgr, ok := c.handlers[name]
	return mgr, ok
}

func (c *AsyncManagerHandlerMap) ListHandlersByName() map[string]AsyncManagerHandler {
	c.lock.RLock()
	defer c.lock.RUnlock()
	result := make(map[string]AsyncManagerHandler)
	for k, v := range c.handlers {
		result[k] = v
	}
	return result
}

func (c *AsyncManagerHandlerMap) RemoveHandler(name string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.handlers, name)
}
