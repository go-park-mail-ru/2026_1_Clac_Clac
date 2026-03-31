package api_test

import (
	"errors"
	"fmt"
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

const (
	StatusOK    = "ok"
	StatusError = "error"
)

func TestSetContentType(t *testing.T) {
	t.Run("just writing", func(t *testing.T) {
		res := httptest.NewRecorder()

		api.SetContentType(res, MIMETextPlain)

		assert.Equal(t, MIMETextPlain, res.Header().Get(HeaderContentType), "types must be equal")
	})

	t.Run("double writing", func(t *testing.T) {
		res := httptest.NewRecorder()

		api.SetContentType(res, MIMETextPlain)
		api.SetContentType(res, MIMEApplicationJSON)

		assert.Equal(t, MIMETextPlain, res.Header().Get(HeaderContentType), "types must be equal")
	})
}

func TestRespond(t *testing.T) {
	t.Run("ok response", func(t *testing.T) {
		expectedBody := fmt.Sprintf(`{"status":"%s"}`, StatusOK)

		res := httptest.NewRecorder()

		_, err := api.Respond(res, http.StatusOK, StatusOK)

		require.NoError(t, err, "respond must not return error")

		assert.Equal(t, MIMEApplicationJSON, res.Header().Get(HeaderContentType), "always response with json")
		assert.Equal(t, http.StatusOK, res.Result().StatusCode, "statuses must be equal")
		assert.Equal(t, expectedBody, res.Body.String(), "bodies must be equal")
	})
}

func TestRespondOk(t *testing.T) {
	t.Run("response with simple user", func(t *testing.T) {
		userId := 5
		userName := "TempName"

		expectedBody := fmt.Sprintf(`{"status":"%s","data":{"id":%d,"name":"%s"}}`, StatusOK, userId, userName)

		simpleUser := struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}{
			ID:   userId,
			Name: userName,
		}

		res := httptest.NewRecorder()

		_, err := api.RespondOk(res, simpleUser)

		require.NoError(t, err, "respond must not return error")

		assert.Equal(t, MIMEApplicationJSON, res.Header().Get(HeaderContentType), "always response with json")
		assert.Equal(t, http.StatusOK, res.Result().StatusCode, "statuses must be equal")
		assert.Equal(t, expectedBody, res.Body.String(), "bodies must be equal")
	})
}

func TestRespondError(t *testing.T) {
	t.Run("error response", func(t *testing.T) {
		errorCode := http.StatusBadRequest
		errorMessage := "this is error message"
		expectedBody := fmt.Sprintf(`{"status":"%s","code":%d,"message":"%s"}`, StatusError, errorCode, errorMessage)

		res := httptest.NewRecorder()

		_, err := api.RespondError(res, http.StatusBadRequest, errorMessage)

		require.NoError(t, err, "respond must not return error")

		assert.Equal(t, MIMEApplicationJSON, res.Header().Get(HeaderContentType), "always response with json")
		assert.Equal(t, http.StatusBadRequest, res.Result().StatusCode, "statuses must be equal")
		assert.Equal(t, expectedBody, res.Body.String(), "bodies must be equal")
	})
}

func TestHandleError(t *testing.T) {
	t.Run("no error test", func(t *testing.T) {
		const zeroStatus = 0

		res := httptest.NewRecorder()
		res.Code = zeroStatus

		err := api.HandleError(res, nil)

		require.NoError(t, err, "must not return error")
		assert.Equal(t, zeroStatus, res.Code, "must not write status code")
		assert.Empty(t, res.Body.String(), "must not write body")
		assert.Empty(t, res.Header(), "must not write header")
	})

	t.Run("error test", func(t *testing.T) {
		res := httptest.NewRecorder()

		err := api.HandleError(res, errors.New("oh no..."))

		require.Error(t, err, "must return error")
		assert.NotEmpty(t, res.Body.String(), "must write body")
		assert.Equal(t, http.StatusInternalServerError, res.Code, "status must be 500")
	})
}
