package service

import (
	"github.com/aakrasnova/zone-mate/storage"
	"github.com/pkg/errors"
)

type Service struct {
	s *storage.Storage
}

func NewService(s *storage.Storage) *Service {
	return &Service{s: s}
}

func (s *Service) AddWithUTCOffset(id int64, offset int) error {
	if offset < -12 || offset > 12 {
		return errors.New("invalid offset, must be in range [-12, 12]")
	}
	return s.s.AddUser(&storage.User{
		ID:       id,
		Timezone: offset,
	})
}

func (s *Service) GetUser(id int64) (*storage.User, error) {
	return s.s.GetUser(id)
}
