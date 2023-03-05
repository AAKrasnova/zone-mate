package storage

import (
	"encoding/binary"
	"encoding/json"

	bolt "go.etcd.io/bbolt"
)

var (
	bktUsers = []byte("users")
)

// Storage is a wrapper around bolt.DB
type Storage struct {
	db *bolt.DB
}

// NewStorage creates a new storage
func NewStorage(path string) (*Storage, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	return &Storage{db: db}, nil
}

// Close closes the storage
func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) AddUser(u *User) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bktUsers)
		if err != nil {
			return err
		}
		encoded, err := encodeUser(u)
		if err != nil {
			return err
		}
		return b.Put(int64ToBytes(u.ID), encoded)
	})
}

func (s *Storage) GetUser(id int64) (*User, error) {
	var u User
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bktUsers)
		if b == nil {
			return nil
		}
		v := b.Get(int64ToBytes(id))
		if v == nil {
			return nil
		}
		err := decodeUser(v, &u)
		if err != nil {
			return err
		}
		return nil
	})
	return &u, err
}

func encodeUser(u *User) ([]byte, error) {
	return json.Marshal(u)
}

func decodeUser(b []byte, u *User) error {
	return json.Unmarshal(b, u)
}

func int64ToBytes(i int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(i))
	return b
}
