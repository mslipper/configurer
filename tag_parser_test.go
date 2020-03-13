package configurer

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTagParser_Parse(t *testing.T) {
	tests := []struct {
		tag string
		out map[string]string
	}{
		{
			"",
			nil,
		},
		{
			"foo,bar",
			map[string]string{
				"foo": "",
				"bar": "",
			},
		},
		{
			"foo",
			map[string]string{
				"foo": "",
			},
		},
		{
			"foo=bar",
			map[string]string{
				"foo": "bar",
			},
		},
		{
			"foo,bar=baz",
			map[string]string{
				"foo": "",
				"bar": "baz",
			},
		},
		{
			"bar=baz,foo",
			map[string]string{
				"foo": "",
				"bar": "baz",
			},
		},
		{
			"foo,bar=baz\\=\\,baz",
			map[string]string{
				"foo": "",
				"bar": "baz=,baz",
			},
		},
		{
			"cab12309u(&*^,reowiguh5897gh",
			map[string]string{
				"cab12309u(&*^":  "",
				"reowiguh5897gh": "",
			},
		},
	}

	for _, tt := range tests {
		parse, err := parseTag(tt.tag)
		require.NoError(t, err)
		require.EqualValues(t, tt.out, parse)
	}
}

func TestTagParser_ParseErrors(t *testing.T) {
	tests := []struct {
		tag string
		err string
	}{
		{
			"=",
			"expected a label",
		},
		{
			",",
			"expected a label",
		},
		{
			"foo=",
			"expected a label",
		},
		{
			"foo=,",
			"expected a label",
		},
		{
			"foo,",
			"expected a label",
		},
		{
			",,",
			"expected a label",
		},
		{
			"foo,,",
			"expected a label",
		},
		{
			"foo=bar=",
			"expected EOF or a separator",
		},
	}

	for _, tt := range tests {
		_, err := parseTag(tt.tag)
		require.Error(t, err)
		require.Contains(t, err.Error(), tt.err)
	}
}
