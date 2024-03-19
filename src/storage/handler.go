package storage

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/pkg/errors"

	"github.com/zeropsio/zcli/src/file"
	"github.com/zeropsio/zcli/src/i18n"
)

type Config struct {
	FilePath string
	FileMode os.FileMode
}

type Handler[T any] struct {
	//nolint:structcheck // Why: `is unused` error is false positive
	config Config
	//nolint:structcheck // Why: `is unused` error is false positive
	data T
	//nolint:structcheck // Why: `is unused` error is false positive
	lock sync.RWMutex
}

func New[T any](config Config) (*Handler[T], error) {
	h := &Handler[T]{
		config: config,
	}

	return h, h.load()
}

func (h *Handler[T]) load() error {
	storageFileExists, err := FileExists(h.config.FilePath)
	if err != nil {
		return errors.WithStack(err)
	}
	if !storageFileExists {
		return nil
	}

	f, err := file.Open(h.config.FilePath, os.O_RDONLY, h.config.FileMode)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()

	// If the file is empty, set the default value and save it.
	fi, err := f.Stat()
	if err != nil {
		return errors.WithStack(err)
	}
	if fi.Size() == 0 {
		return h.Clear()
	}

	if err := json.NewDecoder(f).Decode(&h.data); err != nil {
		return errors.WithMessagef(err, i18n.T(i18n.UnableToDecodeJsonFile, h.config.FilePath))
	}

	return nil
}

func (h *Handler[T]) Clear() error {
	h.lock.Lock()
	defer h.lock.Unlock()
	var data T
	return h.save(data)
}

func (h *Handler[T]) Update(callback func(T) T) (T, error) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.data = callback(h.data)
	return h.data, h.save(h.data)
}

func (h *Handler[T]) save(data T) error {
	h.data = data

	if err := func() error {
		f, err := file.Open(h.config.FilePath+".new", os.O_RDWR|os.O_CREATE|os.O_TRUNC, h.config.FileMode)
		if err != nil {
			return errors.WithStack(err)
		}
		defer f.Close()

		if err := json.NewEncoder(f).Encode(data); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}(); err != nil {
		return err
	}
	if err := os.Rename(h.config.FilePath+".new", h.config.FilePath); err != nil {
		return errors.WithStack(err)
	}
	os.Remove(h.config.FilePath + ".new")
	return nil
}

func (h *Handler[T]) Data() T {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.data
}
