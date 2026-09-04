package config

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

func NewLogger(path string) (*slog.Logger, func(), error) {
	if info, e := os.Stat(path); e == nil && info.Size() > 1024*1024 {
		_ = os.Rename(path, path+".1")
	}
	if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
		return nil, nil, e
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if e != nil {
		return nil, nil, e
	}
	return slog.New(slog.NewJSONHandler(io.MultiWriter(os.Stdout, f), nil)), func() { _ = f.Close() }, nil
}
