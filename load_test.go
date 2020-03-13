package configurer

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type testConfig struct {
	String                string
	Int64                 int64
	Int32                 int32
	Int                   int
	Uint64                uint64
	Uint32                uint32
	Uint                  uint
	Bool                  bool
	DefaultString         string `config:"default=whatever"`
	RequiredString        string `config:"required"`
	DefaultRequiredString string `config:"default=whatever"`
	RequiredRenamedString string `json:"required_renamed_string" toml:"required_renamed_string" yaml:"required_renamed_string" config:"required"`
	Nested                nestedConfig
	NestedPtr             *nestedConfig
	ArrayOfNums           []int
	ArrayOfStuff          []nestedConfig
}

type nestedConfig struct {
	String string `config:"required,default=nested"`
	Bool   bool
}

func TestLoad_ValidConfig_JSON(t *testing.T) {
	abs, err := filepath.Abs("testdata/valid_config.json")
	require.NoError(t, err)
	testValidConfig(t, abs)
}

func TestLoad_ValidConfig_YAML(t *testing.T) {
	abs, err := filepath.Abs("testdata/valid_config.yml")
	require.NoError(t, err)
	testValidConfig(t, abs)
}

func TestLoad_ValidConfig_TOML(t *testing.T) {
	abs, err := filepath.Abs("testdata/valid_config.toml")
	require.NoError(t, err)
	testValidConfig(t, abs)
}

func TestLoad_DefaultValues(t *testing.T) {
	type cfg struct {
		Bool    bool    `config:"default=true"`
		PtrBool *bool   `config:"default=true"`
		Int     int     `config:"default=9"`
		Uint    uint    `config:"default=10"`
		String  string  `config:"default=honkus beepus"`
		Float   float64 `config:"default=10.112"`
		Nested  struct {
			Bool bool `config:"default=true"`
		}
	}
	troo := true
	expCfg := &cfg{
		Bool:    true,
		PtrBool: &troo,
		Int:     9,
		Uint:    10,
		String:  "honkus beepus",
		Float:   10.112,
		Nested: struct {
			Bool bool `config:"default=true"`
		}{
			Bool: true,
		},
	}
	actCfg := new(cfg)
	require.NoError(t, Load(ioutil.NopCloser(bytes.NewReader([]byte("{}"))), DefaultJSONUnmarshaller, actCfg))
	require.EqualValues(t, expCfg, actCfg)
}

func TestLoad_Required(t *testing.T) {
	type cfg struct {
		Bool   bool   `config:"required"`
		Int    int    `config:"required"`
		String string `config:"required"`
		Nested struct {
			Bool bool `config:"required"`
		} `config:"required"`
		Array []int `config:"required"`
	}

	tests := []struct {
		inJSON string
		outErr string
	}{
		{
			"{}",
			"required field Bool not found",
		},
		{
			`{"Bool": null}`,
			"required field Bool is nil",
		},
		{
			`{ "Bool": false }`,
			"required field Int not found",
		},
		{
			`{ "Bool": false, "Int": null }`,
			"required field Int is nil",
		},
		{
			`{ "Bool": false, "Int": 0 }`,
			"required field String not found",
		},
		{
			`{ "Bool": false, "Int": 0, "String": null }`,
			"required field String is nil",
		},
		{
			`{ "Bool": false, "Int": 0, "String": "" }`,
			"required field String is empty",
		},
		{
			`{ "Bool": false, "Int": 0, "String": "test" }`,
			"required field Nested not found",
		},
		{
			`{ "Bool": false, "Int": 0, "String": "test", "Nested": null }`,
			"required field Nested is nil",
		},
		{
			`{ "Bool": false, "Int": 0, "String": "test", "Nested": {} }`,
			"required field Bool not found",
		},
		{
			`{ "Bool": false, "Int": 0, "String": "test", "Nested": { "Bool": null } }`,
			"required field Bool is nil",
		},
		{
			`{ "Bool": false, "Int": 0, "String": "test", "Nested": { "Bool": false } }`,
			"required field Array not found",
		},
		{
			`{ "Bool": false, "Int": 0, "String": "test", "Nested": { "Bool": false }, "Array": null }`,
			"required field Array is nil",
		},
		{
			`{ "Bool": false, "Int": 0, "String": "test", "Nested": { "Bool": false }, "Array": [] }`,
			"required field Array is empty",
		},
	}

	for _, tt := range tests {
		actCfg := new(cfg)
		err := LoadJSON(ioutil.NopCloser(bytes.NewReader([]byte(tt.inJSON))), actCfg)
		require.Error(t, err)
		require.Contains(t, err.Error(), tt.outErr)
	}

	okJSON := `{ "Bool": false, "Int": 0, "String": "test", "Nested": { "Bool": false }, "Array": [ 1 ] }`
	require.NoError(t, LoadJSON(ioutil.NopCloser(bytes.NewReader([]byte(okJSON))), new(cfg)))
}

func TestLoad_DefaultValueOverrides(t *testing.T) {
	type cfg struct {
		String string `config:"default=set from default,env=CONFIGURER_TEST_ENV_VAR"`
	}

	actCfg := new(cfg)
	require.NoError(t, LoadJSON(ioutil.NopCloser(bytes.NewReader([]byte(`{"String": "set in config"}`))), actCfg))
	require.Equal(t, "set in config", actCfg.String)

	actCfg = new(cfg)
	require.NoError(t, LoadJSON(ioutil.NopCloser(bytes.NewReader([]byte(`{}`))), actCfg))
	require.Equal(t, "set from default", actCfg.String)

	require.NoError(t, os.Setenv("CONFIGURER_TEST_ENV_VAR", "set from env var"))
	actCfg = new(cfg)
	require.NoError(t, LoadJSON(ioutil.NopCloser(bytes.NewReader([]byte(`{}`))), actCfg))
	require.Equal(t, "set from env var", actCfg.String)
}

func testValidConfig(t *testing.T, abs string) {
	actCfg := new(testConfig)
	require.NoError(t, LoadURL(fmt.Sprintf("file://%s", abs), actCfg))
	expCfg := &testConfig{
		String:                "hello",
		Int64:                 1,
		Int32:                 2,
		Int:                   3,
		Uint64:                4,
		Uint32:                5,
		Uint:                  6,
		Bool:                  true,
		DefaultString:         "whatever",
		RequiredString:        "required filled in",
		DefaultRequiredString: "whatever",
		RequiredRenamedString: "required_renamed filled in",
		Nested: nestedConfig{
			String: "hello",
			Bool:   false,
		},
		NestedPtr: &nestedConfig{
			String: "nested",
			Bool:   true,
		},
		ArrayOfNums: []int{
			1,
			2,
			3,
		},
		ArrayOfStuff: []nestedConfig{
			{
				String: "nested",
				Bool:   false,
			},
			{
				String: "what is up again",
				Bool:   true,
			},
		},
	}
	require.EqualValues(t, expCfg, actCfg)
}
