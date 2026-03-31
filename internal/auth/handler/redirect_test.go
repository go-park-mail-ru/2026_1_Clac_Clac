package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedirect(t *testing.T) {
	const codeParam = "code"
	const messageParam = "message"

	tests := []struct {
		Name            string
		TargetURL       string
		ExpectedCode    int
		ExpectedMessage string
		ExpectError     bool
	}{
		{
			Name:            "no error, success redirect",
			TargetURL:       "/",
			ExpectedCode:    200,
			ExpectedMessage: "success",
			ExpectError:     false,
		},
		{
			Name:            "error, invalid target url",
			TargetURL:       "\n",
			ExpectedCode:    0,
			ExpectedMessage: "",
			ExpectError:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			res := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/callback", nil)

			w, err := handler.Redirect(res, req, test.TargetURL, test.ExpectedCode, test.ExpectedMessage)
			require.Equal(t, res, w, "responses must be equal")
			if test.ExpectError {
				require.Error(t, err, "redirect must return error")
				return
			}
			require.NoError(t, err, "redirect must not return error")

			r := res.Result()
			assert.Equal(t, http.StatusFound, r.StatusCode, "status must be 302")

			location, err := r.Location()
			require.NoError(t, err, "location must exists")

			values := location.Query()

			require.True(t, values.Has(codeParam), "must contain code in query params")
			assert.Equal(t, strconv.Itoa(test.ExpectedCode), values.Get(codeParam), "codes must be equal")

			require.True(t, values.Has(messageParam), "must contain message in query params")
			assert.Equal(t, test.ExpectedMessage, values.Get(messageParam), "messages must be equal")
		})
	}
}
