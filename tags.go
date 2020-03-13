package configurer

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"os"
	"reflect"
	"strings"
)

const (
	TagName = "config"
)

type FieldConfig struct {
	Required bool
	Default  string
	Env      string
	Defined  bool
}

func processTags(v interface{}, unmarshaller Unmarshaller, keyMap map[string]interface{}) error {
	cfgVal := reflect.Indirect(reflect.ValueOf(v))
	if cfgVal.Kind() == reflect.Interface {
		cfgVal = cfgVal.Elem()
	}
	cfgKind := cfgVal.Kind()
	if cfgKind != reflect.Struct {
		return fmt.Errorf("can only process structs, but got %s", cfgKind.String())
	}

	cfgType := cfgVal.Type()
	for i := 0; i < cfgType.NumField(); i++ {
		fieldDef := cfgType.Field(i)
		fieldVal := cfgVal.Field(i)
		fieldCfg, err := parseStructTag(fieldDef.Tag.Get(TagName))
		if err != nil {
			return errors.Wrap(err, "invalid configure struct tag")
		}

		var envOverride string
		if fieldCfg.Env != "" {
			envOverride, _ = os.LookupEnv(fieldCfg.Env)
		}

		fieldName := strings.ToLower(unmarshaller.ExtractFieldName(fieldDef))
		rawFieldVal, rawFieldIsDefined := keyMap[fieldName]
		rawFieldIsNil := rawFieldIsDefined && rawFieldVal == nil
		hasDefault := envOverride != "" || fieldCfg.Default != ""
		derefFieldVal := fieldVal
		for derefFieldVal.Kind() == reflect.Ptr {
			derefFieldVal = derefFieldVal.Elem()
		}

		derefFieldValKind := derefFieldVal.Kind()
		if hasDefault && derefFieldValKind == reflect.Struct {
			return fmt.Errorf("struct field %s cannot have a default defined", fieldDef.Name)
		}

		var appliedDefault bool
		if rawFieldVal == nil || !rawFieldIsDefined {
			if envOverride != "" {
				if err := yaml.Unmarshal([]byte(envOverride), fieldVal.Addr().Interface()); err != nil {
					return errors.Wrap(err, fmt.Sprintf("couldn't unmarshal env var %s into field %s", fieldCfg.Env, fieldDef.Name))
				}
				appliedDefault = true
			} else if fieldCfg.Default != "" {
				if err := yaml.Unmarshal([]byte(fieldCfg.Default), fieldVal.Addr().Interface()); err != nil {
					return errors.Wrap(err, fmt.Sprintf("couldn't unmarshal default value for field %s", fieldDef.Name))
				}
				appliedDefault = true
			}
		}

		if !rawFieldIsDefined && !appliedDefault && fieldCfg.Required {
			return fmt.Errorf("required field %s not found", fieldDef.Name)
		}

		if rawFieldIsNil && !appliedDefault && fieldCfg.Required {
			return fmt.Errorf("required field %s is nil", fieldDef.Name)
		}

		if derefFieldValKind == reflect.String {
			strLen := derefFieldVal.Len()
			if fieldCfg.Required && strLen == 0 {
				return fmt.Errorf("required field %s is empty", fieldDef.Name)
			}
		}

		if derefFieldValKind == reflect.Slice {
			sliceLen := derefFieldVal.Len()
			if fieldCfg.Required && sliceLen == 0 {
				return fmt.Errorf("required field %s is empty", fieldDef.Name)
			}

			elemType := derefFieldVal.Type().Elem()
			if elemType.Kind() != reflect.Struct {
				continue
			}

			nextKeyMapVal := reflect.ValueOf(rawFieldVal)
			for i := 0; i < sliceLen; i++ {
				next := nextKeyMapVal.Index(i).Interface().(map[string]interface{})
				if err := processTags(derefFieldVal.Index(i).Addr().Interface(), unmarshaller, next); err != nil {
					return errors.Wrap(err, fmt.Sprintf("error processing array field %s", fieldDef.Name))
				}
			}
			continue
		}

		if derefFieldValKind == reflect.Struct {
			if rawFieldVal == nil {
				if err := processTags(derefFieldVal.Addr().Interface(), unmarshaller, make(map[string]interface{})); err != nil {
					return errors.Wrap(err, fmt.Sprintf("error processing field %s", fieldDef.Name))
				}
				continue
			}

			if err := processTags(derefFieldVal.Addr().Interface(), unmarshaller, rawFieldVal.(map[string]interface{})); err != nil {
				return errors.Wrap(err, fmt.Sprintf("error processing field %s", fieldDef.Name))
			}
		}
	}

	return nil
}

func parseStructTag(tag string) (*FieldConfig, error) {
	cfg := new(FieldConfig)
	if tag == "" {
		return cfg, nil
	}

	parsed, err := parseTag(tag)
	if err != nil {
		return nil, errors.Wrap(err, "mal-formed tag")
	}

	cfg.Defined = true
	_, cfg.Required = parsed["required"]
	cfg.Default = parsed["default"]
	cfg.Env = parsed["env"]
	return cfg, nil
}
