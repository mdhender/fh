// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package sessions

import (
	"encoding/json"
	"github.com/google/uuid"
	"os"
	"sync"
	"time"
)

//func Load(path string) (*Store, error) {
//	store := &Store{
//		sessions: make(map[string]Session),
//	}
//	if path == ":memory:" {
//		return store, nil
//	}
//	data, err := os.ReadFile(path)
//	if err != nil {
//		return nil, err
//	} else if err = json.Unmarshal(data, store); err != nil {
//		return nil, err
//	}
//	return store, nil
//}
//
//func Save(s *Store, path string) error {
//	s.Lock()
//	defer s.Unlock()
//	data, err := json.Marshal(s)
//	if err != nil {
//		return err
//	}
//	return os.WriteFile(path, data, 0644)
//}

// Manager is a session manager and factory
type Manager struct {
	sync.Mutex
	Sessions map[string]Session
	Ttl      time.Duration
	signing  struct {
		key          []byte
		salt, pepper []byte
	}
}

func NewManager(ttl time.Duration) *Manager {
	return &Manager{
		Sessions: make(map[string]Session),
		Ttl:      ttl,
	}
}

func (m *Manager) Create(acctId string) Session {
	id, err := uuid.NewRandom()
	if err != nil {
		return Session{}
	}
	sess := Session{
		Id:        id.String(),
		Account:   acctId,
		ExpiresAt: time.Now().Add(m.Ttl),
	}

	m.Lock()
	defer m.Unlock()

	m.Sessions[sess.Id] = sess

	return sess
}

func (m *Manager) Delete(id string) {
	m.Lock()
	defer m.Unlock()
	delete(m.Sessions, id)
}

func (m *Manager) Fetch(id string) (Session, bool) {
	m.Lock()
	defer m.Unlock()

	sess, ok := m.Sessions[id]
	if !ok {
		return Session{}, false
	} else if sess.IsExpired() {
		delete(m.Sessions, id)
	}
	return Session{}, false
}

func (m *Manager) Load(path string) error {
	m.Lock()
	defer m.Unlock()

	m.Sessions = make(map[string]Session)

	var store []Session
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	} else if err = json.Unmarshal(data, &store); err != nil {
		return err
	}

	for _, sess := range store {
		m.Sessions[sess.Id] = sess
	}

	return nil
}

func (m *Manager) Save(path string) error {
	m.Lock()
	defer m.Unlock()

	var store []Session
	for _, sess := range m.Sessions {
		store = append(store, sess)
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	} else if err = os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	return nil
}

// Session is a session
type Session struct {
	Id        string
	Account   string
	ExpiresAt time.Time
}

func (s Session) IsExpired() bool {
	return !time.Now().Before(s.ExpiresAt)
}

func (s Session) IsValid() bool {
	return time.Now().Before(s.ExpiresAt)
}
