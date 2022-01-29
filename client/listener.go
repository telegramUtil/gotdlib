package client

import (
	"sync"
)

func newListenerStore() *listenerStore {
	return &listenerStore{
		listeners: []*Listener{},
	}
}

type listenerStore struct {
	sync.Mutex
	listeners []*Listener
}

func (store *listenerStore) Add(listener *Listener) {
	store.Lock()
	defer store.Unlock()

	store.listeners = append(store.listeners, listener)
}

func (store *listenerStore) Listeners() []*Listener {
	store.Lock()
	defer store.Unlock()

	return store.listeners
}

func (store *listenerStore) gc() {
	store.Lock()
	defer store.Unlock()

	oldListeners := store.listeners

	store.listeners = []*Listener{}

	for _, listener := range oldListeners {
		if listener.IsActive() {
			store.listeners = append(store.listeners, listener)
		}
	}
}

type Listener struct {
	mu         sync.Mutex
	isActive   bool
	Updates    chan Type
	RawUpdates chan Type
	Filter     Type
}

func (listener *Listener) Close() {
	listener.mu.Lock()
	defer listener.mu.Unlock()

	listener.isActive = false
	if listener.Updates != nil {
		close(listener.Updates)
	}
	if listener.RawUpdates != nil {
		close(listener.RawUpdates)
	}
}

func (listener *Listener) IsActive() bool {
	listener.mu.Lock()
	defer listener.mu.Unlock()

	return listener.isActive
}
