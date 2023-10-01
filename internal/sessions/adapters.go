// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package sessions

import "time"

func AdaptStoreToJSONStore(s *Store) (*JSONStore, error) {
	s.Lock()
	defer s.Unlock()

	var js JSONStore
	js.Sessions = make(map[string]JSONSession)
	js.Signing.Salt = s.signing.salt
	js.Signing.Key = s.signing.key
	js.Signing.Pepper = s.signing.pepper
	for _, sess := range s.sessions {
		if !sess.IsExpired() {
			js.Sessions[sess.Id] = JSONSession{
				Account:   sess.Account,
				ExpiresAt: sess.ExpiresAt.Format(time.RFC3339),
			}
		}
	}

	return &js, nil
}

func AdaptJSONStoreToStore(js *JSONStore) (*Store, error) {
	s := &Store{sessions: make(map[string]Session)}
	s.signing.salt = js.Signing.Salt
	s.signing.key = js.Signing.Key
	s.signing.pepper = js.Signing.Pepper

	for id, ss := range js.Sessions {
		exp, err := time.Parse(time.RFC3339, ss.ExpiresAt)
		if err != nil {
			return nil, err
		}
		s.sessions[id] = Session{
			Id:        id,
			Account:   ss.Account,
			ExpiresAt: exp,
		}
	}

	return s, nil
}
