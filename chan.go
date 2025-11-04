package nlvMisc

import "sync"

const CHANBUFFSIZE = 32

// MultiChan represents a thread-safe interface for managing multiple channels of type T.
// It supports adding channels, broadcasting data to all channels, and closing all channels.
type MultiChan[T any] interface {
	AddChan() <-chan T
	Close()
	Send(T)
	Len() int
	IsActive() bool
}

// multipleChannel is a thread-safe structure for managing a collection of channels of type T.
// It supports adding channels, broadcasting data, and safely closing all channels.
type multipleChannel[T any] struct {
	channelList []chan T
	lock        sync.Mutex
	active      bool
}

// NewMultiChan initializes and returns a new MultiChan instance with the specified capacity.
func NewMultiChan[T any](capacity int) MultiChan[T] {
	var m multipleChannel[T]
	if capacity <= 0 {
		capacity = CHANBUFFSIZE
	} // alternative: panic?
	m.channelList = make([]chan T, 0, capacity)
	m.active = true
	return &m
}

// IsActive returns true if the multipleChannel is active and
// accepting operations, false if it has been closed.
func (m *multipleChannel[T]) IsActive() bool {
	return m.active
}

// AddChan adds a new channel to the MultiChan and returns the receiver channel
func (m *multipleChannel[T]) AddChan() <-chan T {
	ch := make(chan T)
	m.lock.Lock()
	defer m.lock.Unlock()
	if !m.active {
		panic("attempt to add new channel to closed MultiChan")
	}
	m.channelList = append(m.channelList, ch)
	return ch
}

// Close shuts down all channels managed by the MultiChan
// permanently, release the channellist, and marks the object
// as terminated (inactive).
func (m *multipleChannel[T]) Close() {
	m.lock.Lock()
	defer m.lock.Unlock()
	if !m.active {
		panic("attempt to close a MultiChan that has already been closed")
	}
	for _, ch := range m.channelList {
		close(ch)
	}
	m.channelList = nil
	m.active = false
}

// Send sends the provided value `t` to all channels managed by the MultiChan.
// It locks the MultiChan to ensure thread-safe operations.
func (m *multipleChannel[T]) Send(t T) {
	m.lock.Lock()
	defer m.lock.Unlock()
	switch {
	case !m.active:
		panic("attempt to send to a MultiChan that has already been closed")
	case 0 == len(m.channelList) || nil == m.channelList:
		panic("attempt to send to a MultiChan with no channels")
	default:
		for _, ch := range m.channelList {
			ch <- t
		}
	}
}

// Len returns the number of channels currently managed by the MultiChan.
// Use of append() when adding channels means there is a potential race
// condition / invalid access, so must lock prior to testing
func (m *multipleChannel[T]) Len() int {
	m.lock.Lock()
	defer m.lock.Unlock()
	if !m.active {
		panic("attempt to count channels in a MultiChan that has already been closed")
	}
	return len(m.channelList)
}
