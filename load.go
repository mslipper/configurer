package configurer

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"strings"
)

type Loader struct {
	sources      map[string]Source
	unmarshalers map[string]Unmarshaller
}

func NewLoader() *Loader {
	return &Loader{
		sources:      make(map[string]Source),
		unmarshalers: make(map[string]Unmarshaller),
	}
}

func (l *Loader) LoadURL(url string, v interface{}) error {
	protoIdx := strings.Index(url, "://")
	if protoIdx == -1 {
		return errors.New("url should start with some protocol")
	}
	proto := url[:protoIdx]
	source := l.sources[proto]
	if source == nil {
		return fmt.Errorf("can't find source for protocol %s - try registering one", proto)
	}

	extIdx := strings.LastIndex(url, ".")
	if extIdx == -1 {
		return errors.New("url does not have a detectable file extension. please either add one or force a config type")
	}
	if extIdx == len(url)-1 {
		return errors.New("url ends in a dot, but doesn't have a detectable file extension. please either add one or force a config type")
	}
	ext := url[extIdx+1:]

	unmarshaller := l.unmarshalers[ext]
	if unmarshaller == nil {
		return fmt.Errorf("can't find unmarshaller for extension %s - try registering one", ext)
	}

	r, err := source.Reader(url)
	if err != nil {
		return errors.Wrap(err, "error opening config")
	}
	return l.Load(r, unmarshaller, v)
}

func (l *Loader) LoadJSON(r io.ReadCloser, v interface{}) error {
	return l.Load(r, l.unmarshalers[JSON], v)
}

func (l *Loader) LoadTOML(r io.ReadCloser, v interface{}) error {
	return l.Load(r, l.unmarshalers[TOML], v)
}

func (l *Loader) LoadYAML(r io.ReadCloser, v interface{}) error {
	return l.Load(r, l.unmarshalers[YAML], v)
}

func (l *Loader) Load(r io.ReadCloser, unmarshaller Unmarshaller, v interface{}) error {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return errors.Wrap(err, "error reading config")
	}
	if err := r.Close(); err != nil {
		return errors.Wrap(err, "error closing reader")
	}
	if err := unmarshaller.Unmarshal(buf, v); err != nil {
		return errors.Wrap(err, "error unmarshalling config")
	}
	keyMap := make(map[string]interface{})
	if err := unmarshaller.Unmarshal(buf, &keyMap); err != nil {
		return errors.Wrap(err, "error unmarshalling config")
	}
	return processTags(v, unmarshaller, l.lowercaseKeyMap(keyMap))
}

func (l *Loader) lowercaseKeyMap(m map[string]interface{}) map[string]interface{} {
	return l.handleMapValue(m).(map[string]interface{})
}

func (l *Loader) handleMapValue(v interface{}) interface{} {
	switch v := v.(type) {
	case map[string]interface{}:
		return l.handleMap(v)
	case []interface{}:
		return l.handleSlice(v)
	default:
		return v
	}
}

func (l *Loader) handleMap(in map[string]interface{}) interface{} {
	res := make(map[string]interface{})
	for k, v := range in {
		res[strings.ToLower(k)] = l.handleMapValue(v)
	}
	return res
}

func (l *Loader) handleSlice(in []interface{}) []interface{} {
	res := make([]interface{}, len(in))
	for i, v := range in {
		res[i] = l.handleMapValue(v)
	}
	return res
}

func (l *Loader) RegisterSource(source Source) {
	protos := source.Protocols()
	for _, proto := range protos {
		if l.sources[proto] != nil {
			panic(fmt.Sprintf("source with protocol %s already registered", proto))
		}
		l.sources[proto] = source
	}
}

func (l *Loader) RegisterUnmarshaller(unmarshaller Unmarshaller) {
	exts := unmarshaller.Extensions()
	for _, ext := range exts {
		if l.unmarshalers[ext] != nil {
			panic(fmt.Sprintf("unmarshaller with extension %s already registered", ext))
		}
		l.unmarshalers[ext] = unmarshaller
	}
}

var defaultLoader = NewLoader()

func LoadURL(url string, v interface{}) error {
	return defaultLoader.LoadURL(url, v)
}

func LoadJSON(r io.ReadCloser, v interface{}) error {
	return defaultLoader.LoadJSON(r, v)
}

func LoadTOML(r io.ReadCloser, v interface{}) error {
	return defaultLoader.LoadTOML(r, v)
}

func LoadYAML(r io.ReadCloser, v interface{}) error {
	return defaultLoader.LoadYAML(r, v)
}

func Load(r io.ReadCloser, unmarshaller Unmarshaller, v interface{}) error {
	return defaultLoader.Load(r, unmarshaller, v)
}

func RegisterSource(source Source) {
	defaultLoader.RegisterSource(source)
}

func RegisterUnmarshaller(unmarshaller Unmarshaller) {
	defaultLoader.RegisterUnmarshaller(unmarshaller)
}
