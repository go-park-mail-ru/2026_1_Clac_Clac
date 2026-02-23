package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Точно такие же константы в api, сделал это осознано.
// Логика такая: код должен соответсовать тестам,
// если брать константы из api, то может быть опечатка.
// Вероятность опечататся дважды ниже, чем один
const (
	HeaderContentType = "Content-Type"
)

const (
	MIMEApplicationJSON = "application/json"
	MIMETextPlain       = "text/plain"
)

func TestRespondString(t *testing.T) {
	res := httptest.NewRecorder()
	content := "Hello, World!"

	err := api.RespondString(res, http.StatusOK, content)
	require.NoError(t, err)

	header := res.Header()

	assert.Equal(t, MIMETextPlain, header.Get(HeaderContentType), "mime types dont equal")
	assert.Equal(t, content, res.Body.String(), "contents are different")
}

func TestRespondJSON(t *testing.T) {
	res := httptest.NewRecorder()
	testStruct := struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}{
		ID:   5,
		Name: "User",
	}
	want, err := json.Marshal(testStruct)
	require.NoError(t, err, "incorrect test struct")

	err = api.RespondJSON(res, http.StatusOK, testStruct)
	require.NoError(t, err)

	header := res.Header()

	assert.Equal(t, MIMEApplicationJSON, header.Get(HeaderContentType), "mime types dont equal")
	assert.Equal(t, want, res.Body.Bytes(), "contents are different")
}
