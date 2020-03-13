package configurer

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"os"
	"strings"
)

type Source interface {
	Protocols() []string
	Reader(url string) (io.ReadCloser, error)
}

type FileSource struct {
}

func (f *FileSource) Protocols() []string {
	return []string{"file"}
}

func (f *FileSource) Reader(url string) (io.ReadCloser, error) {
	path := strings.TrimPrefix(url, "file://")
	return os.OpenFile(path, os.O_RDONLY, 0)
}

type HTTPSource struct {
}

func (h *HTTPSource) Protocols() []string {
	return []string{"http", "https"}
}

func (h *HTTPSource) Reader(url string) (io.ReadCloser, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "error getting URL")
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("expected 200 response code but got %d", res.StatusCode)
	}
	return res.Body, nil
}

func init() {
	RegisterSource(new(FileSource))
	RegisterSource(new(HTTPSource))
}
