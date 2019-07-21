package properties

import (
	"container/list"
	"sync"
)

// sequence hashtable
type sequenceTable struct {
	mutex   sync.Mutex
	mapper  map[interface{}]interface{}
	element map[interface{}]*list.Element
	list    *list.List
}

func NewHashtable2() Hashtable {
	return &sequenceTable{
		mapper:  map[interface{}]interface{}{},
		element: map[interface{}]*list.Element{},
		list:    list.New(),
	}
}

func (s *sequenceTable) New() Hashtable {
	return NewHashtable2()
}

func (s *sequenceTable) Put(key, value interface{}) interface{} {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	old := s.mapper[key]
	s.mapper[key] = value
	s.element[key] = s.list.PushBack(key)

	return old
}

func (s *sequenceTable) Get(key interface{}) interface{} {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.mapper[key]
}

func (s *sequenceTable) Remove(key interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.mapper, key)
	s.list.Remove(s.element[key])
	delete(s.element, key)
}

func (s *sequenceTable) Size() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.list.Len()
}

func (s *sequenceTable) Keys() []interface{} {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var keys = make([]interface{}, 0, s.list.Len())
	for node := s.list.Front(); node != nil; node = node.Next() {
		keys = append(keys, node.Value)
	}

	return keys
}

// sequence properties
type Properties2 struct {
	Properties
}

func NewProperties2() *Properties2 {
	properties := NewProperties()
	properties.Hashtable = NewHashtable2()

	return &Properties2{
		Properties: *properties,
	}
}
