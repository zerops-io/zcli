package storage

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/zerops-io/zcli/src/utils"
)

type Config struct {
	FilePath string
}

type Handler struct {
	config Config

	lock sync.Mutex
}

func New(config Config) (*Handler, error) {
	h := &Handler{
		config: config,
	}

	dir, _ := path.Split(config.FilePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(config.FilePath, os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return h, nil
}

func (h *Handler) Load(data interface{}) interface{} {
	h.lock.Lock()
	defer h.lock.Unlock()

	if utils.FileExists(h.config.FilePath) {
		data, err := func() (interface{}, error) {
			f, err := os.Open(h.config.FilePath)
			if err != nil {
				return nil, err
			}
			defer f.Close()

			bytes, err := ioutil.ReadAll(f)
			if err != nil {
				return nil, err
			}

			err = json.Unmarshal(bytes, &data)
			if err != nil {
				return nil, err
			}

			return data, nil
		}()
		if err == nil {
			return data
		}
	}

	return data
}

func (h *Handler) Save(data interface{}) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(h.config.FilePath, dataBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}
