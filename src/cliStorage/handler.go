package cliStorage

import (
	"github.com/zerops-io/zcli/src/utils/storage"
)

type Config struct {
	FilePath string
}

type Handler struct {
	storage *storage.Handler
}

type Data struct {
	ProjectId string
	ServerIp  string
	Token     string
}

func New(config Config) (*Handler, error) {
	s, err := storage.New(storage.Config{
		FilePath: config.FilePath,
	})
	if err != nil {
		return nil, err
	}

	h := &Handler{
		storage: s,
	}

	return h, nil
}

func (h *Handler) Data() *Data {
	data := h.storage.Load(&Data{})
	if d, ok := data.(*Data); ok {
		return d
	}
	return nil
}

func (h *Handler) Save(data *Data) error {
	return h.storage.Save(data)
}
