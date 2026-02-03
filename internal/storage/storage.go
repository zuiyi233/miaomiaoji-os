package storage

import (
	"io"
)

type Storage interface {
	Put(key string, reader io.Reader, contentType string) error
	Open(key string) (io.ReadCloser, error)
	Delete(key string) error
	Exists(key string) (bool, error)
}
