package configurer

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHTTPSource(t *testing.T) {
	source := new(HTTPSource)
	require.EqualValues(t, []string{"http", "https"}, source.Protocols())

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "testing")
	}))
	defer ts.Close()

	rd, err := source.Reader(ts.URL)
	require.NoError(t, err)
	data, err := ioutil.ReadAll(rd)
	require.NoError(t, err)
	require.Equal(t, "testing", string(data))
}

func TestFileSource(t *testing.T) {
	source := new(FileSource)
	tmp, err := ioutil.TempFile("", "configurer_")
	require.NoError(t, err)
	content := "this file contains configurer test content"
	_, err = tmp.Write([]byte(content))
	require.NoError(t, err)

	rd, err := source.Reader(fmt.Sprintf("file://%s", tmp.Name()))
	require.NoError(t, err)
	data, err := ioutil.ReadAll(rd)
	require.NoError(t, err)
	require.Equal(t, content, string(data))
	require.NoError(t, rd.Close())
	require.NoError(t, os.Remove(tmp.Name()))
}
