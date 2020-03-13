package configurer

import (
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
	"reflect"
	"strings"
)

type Unmarshaller interface {
	Extensions() []string
	ExtractFieldName(field reflect.StructField) string
	Unmarshal(data []byte, v interface{}) error
}

const (
	TOML = "toml"
	JSON = "json"
	YAML = "yaml"
)

type TOMLUnmarshaller struct {
}

var DefaultTOMLUnmarshaller = new(TOMLUnmarshaller)

func (t *TOMLUnmarshaller) Extensions() []string {
	return []string{TOML}
}

func (t *TOMLUnmarshaller) ExtractFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("toml")
	if tag == "" {
		return field.Name
	}
	return tag
}

func (t *TOMLUnmarshaller) Unmarshal(data []byte, v interface{}) error {
	_, err := toml.Decode(string(data), v)
	return err
}

type JSONUnmarshaller struct {
}

var DefaultJSONUnmarshaller = new(JSONUnmarshaller)

func (j *JSONUnmarshaller) Extensions() []string {
	return []string{JSON}
}

func (j *JSONUnmarshaller) ExtractFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return field.Name
	}
	tagFields := strings.Split(tag, ",")
	if tagFields[0] == "" {
		return field.Name
	}
	return tagFields[0]
}

func (j *JSONUnmarshaller) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

type YAMLUnmarshaller struct {
}

var DefaultYAMLUnmarshaller = new(YAMLUnmarshaller)

func (y *YAMLUnmarshaller) Extensions() []string {
	return []string{YAML, "yml"}
}

func (y *YAMLUnmarshaller) ExtractFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("yaml")
	if tag == "" {
		return field.Name
	}
	tagFields := strings.Split(tag, ",")
	if tagFields[0] == "" {
		return field.Name
	}
	return tagFields[0]
}

// see https://github.com/go-yaml/yaml/issues/139
func (y *YAMLUnmarshaller) Unmarshal(data []byte, v interface{}) error {
	switch v := v.(type) {
	case *map[string]interface{}:
		var res interface{}
		if err := yaml.Unmarshal(data, &res); err != nil {
			return err
		}
		mapVal := y.cleanupMapValue(res).(map[string]interface{})
		*v = mapVal
		return nil
	default:
		return yaml.Unmarshal(data, v)
	}
}

func (y *YAMLUnmarshaller) cleanupMapValue(v interface{}) interface{} {
	switch v := v.(type) {
	case []interface{}:
		return y.cleanupInterfaceArray(v)
	case map[interface{}]interface{}:
		return y.cleanupInterfaceMap(v)
	default:
		return v
	}
}

func (y *YAMLUnmarshaller) cleanupInterfaceArray(in []interface{}) []interface{} {
	res := make([]interface{}, len(in))
	for i, v := range in {
		res[i] = y.cleanupMapValue(v)
	}
	return res
}

func (y *YAMLUnmarshaller) cleanupInterfaceMap(in map[interface{}]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range in {
		res[fmt.Sprintf("%v", k)] = y.cleanupMapValue(v)
	}
	return res
}

func init() {
	RegisterUnmarshaller(DefaultTOMLUnmarshaller)
	RegisterUnmarshaller(DefaultJSONUnmarshaller)
	RegisterUnmarshaller(DefaultYAMLUnmarshaller)
}
