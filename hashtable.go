package properties

import "sync"

type Hashtable interface {
	New() Hashtable
	Put(key, value interface{}) interface{}
	Get(key interface{}) interface{}
	Remove(key interface{})
	Size() int
	Keys() []interface{}
}

type hashtable struct {
	mutex  sync.RWMutex
	mapper map[interface{}]interface{}
}

func NewHashtable() Hashtable {
	return &hashtable{
		mapper: map[interface{}]interface{}{},
	}
}

func (h *hashtable) New() Hashtable {
	return NewHashtable()
}

func (h *hashtable) Put(key, value interface{}) interface{} {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	var old = h.mapper[key]
	h.mapper[key] = value

	return old
}

func (h *hashtable) Get(key interface{}) interface{} {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.mapper[key]
}

func (h *hashtable) Remove(key interface{}) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	delete(h.mapper, key)
}

func (h *hashtable) Size() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return len(h.mapper)
}

func (h *hashtable) Keys() []interface{} {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	var keys = make([]interface{}, 0, len(h.mapper))
	for key := range h.mapper {
		keys = append(keys, key)
	}

	return keys
}
