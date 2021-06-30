package utils

import (
	"fmt"
	"sync"

	"github.com/saveio/themis/common"
)

type Set struct {
	items map[interface{}]struct{}
	lock  sync.RWMutex
}

func NewSet() *Set {
	var set Set
	set.items = make(map[interface{}]struct{})
	return &set
}

func (set *Set) Add(items ...interface{}) {
	set.lock.Lock()
	defer set.lock.Unlock()

	for _, item := range items {
		set.items[item] = struct{}{}
	}
}

func (set *Set) Remove(items ...interface{}) {
	set.lock.Lock()
	defer set.lock.Unlock()

	for _, item := range items {
		delete(set.items, item)
	}
}

func (set *Set) Exists(item interface{}) bool {
	set.lock.RLock()
	defer set.lock.RUnlock()

	_, ok := set.items[item]
	return ok
}

func (set *Set) Len() int64 {
	set.lock.RLock()
	defer set.lock.RUnlock()

	size := int64(len(set.items))
	return size
}

func (set *Set) Clear() {
	set.lock.Lock()
	defer set.lock.Unlock()

	set.items = map[interface{}]struct{}{}
}

func (set *Set) GetAllAddrs() []common.Address {
	set.lock.RLock()
	defer set.lock.RUnlock()

	var addrs []common.Address
	for k := range set.items {
		item := k.(common.Address)
		addrs = append(addrs, item)
	}
	return addrs
}

func (set *Set) PrintAllAddrs() {
	set.lock.RLock()
	defer set.lock.RUnlock()
	fmt.Println("PrintAllAddrs: ")
	for k := range set.items {
		item := k.(common.Address)
		fmt.Println("Print item: ", item)
	}
}

func (set *Set) AddrSerialize() ([]byte, error) {
	var data []byte
	for k := range set.items {
		item := k.(common.Address)
		bItem := item[0:20]
		data = append(data, bItem...)
	}
	return data, nil
}

func (set *Set) AddrDeserialize(data []byte) error {
	dataLen := len(data)
	addrLen := len(common.Address{})
	var tmpAddr common.Address
	for i := 0; i < dataLen; {
		copy(tmpAddr[:], data[i:i+addrLen])
		set.items[tmpAddr] = struct{}{}
		i = i + addrLen
	}
	return nil
}
