package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/api"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	authServiceMocks "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/handler/auth/mock_auth_srv"
	vkOAuthMocks "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/handler/auth/mock_vk_oauth"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
)

type mockTransport struct {
	RoundTripFunc func(req *http.Request) *http.Response
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req), nil
}

func TestVkOAuthCallbackExistingUser(t *testing.T) {
	mockVkOAuth := new(vkOAuthMocks.VkOAuth)
	mockAuthService := new(authServiceMocks.AuthService)

	handler := &AuthHandler{srv: mockAuthService}

	conf := &config.VkOAuth{APIMethod: "https://api.vk.com/method/users.get?access_token=%s"}
	redirectTo := "/"
	testEmail := "user@example.com"
	testToken := &oauth2.Token{AccessToken: "fake-token"}
	testToken = testToken.WithExtra(map[string]any{"email": testEmail})

	mockClient := &http.Client{
		Transport: &mockTransport{
			RoundTripFunc: func(req *http.Request) *http.Response {
				vkResp := api.VkAPIUsersData{
					Response: []api.VkAPIUserData{{FirstName: "Ivan"}},
				}
				body, _ := json.Marshal(vkResp)

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBuffer(body)),
					Header:     make(http.Header),
				}
			},
		},
	}

	mockVkOAuth.On("Exchange", mock.Anything, "valid_code").Return(testToken, nil)
	mockVkOAuth.On("Client", mock.Anything, testToken).Return(mockClient)

	mockAuthService.On("GetUserByEmail", mock.Anything, testEmail).
		Return(models.User{ID: common.FixedUserUuiD, Email: testEmail}, nil)

	mockAuthService.On("CreateSessionForUser", mock.Anything, mock.AnythingOfType("models.User")).
		Return("fake-session-id", nil)

	req := httptest.NewRequest(http.MethodGet, "/callback?code=valid_code", nil)
	rr := httptest.NewRecorder()

	httpHandler := handler.VkOAuthCallback(conf, redirectTo, mockVkOAuth)
	httpHandler(rr, req)

	assert.Equal(t, http.StatusFound, rr.Code)
	assert.Equal(t, redirectTo, rr.Header().Get("Location"))

	cookies := rr.Result().Cookies()
	assert.NotEmpty(t, cookies)
	assert.Equal(t, "fake-session-id", cookies[0].Value)

	mockVkOAuth.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestVkOAuthCallbackNewUserRegistration(t *testing.T) {
	mockVkOAuth := new(vkOAuthMocks.VkOAuth)
	mockAuthService := new(authServiceMocks.AuthService)

	handler := &AuthHandler{srv: mockAuthService}

	conf := &config.VkOAuth{APIMethod: "https://api.vk.com/method/users.get?access_token=%s"}
	redirectTo := "/"
	testEmail := "new@example.com"
	testToken := &oauth2.Token{AccessToken: "fake-token"}
	testToken = testToken.WithExtra(map[string]any{"email": testEmail})

	mockClient := &http.Client{
		Transport: &mockTransport{
			RoundTripFunc: func(req *http.Request) *http.Response {
				vkResp := api.VkAPIUsersData{
					Response: []api.VkAPIUserData{{FirstName: "Alice"}},
				}
				body, _ := json.Marshal(vkResp)
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer(body))}
			},
		},
	}

	mockVkOAuth.On("Exchange", mock.Anything, "valid_code").Return(testToken, nil)
	mockVkOAuth.On("Client", mock.Anything, testToken).Return(mockClient)

	mockAuthService.On("GetUserByEmail", mock.Anything, testEmail).
		Return(models.User{}, common.ErrorNonexistentUser)

	mockAuthService.On("Register", mock.Anything, "Alice", mock.AnythingOfType("string"), testEmail).
		Return(models.User{ID: common.FixedUserUuiD}, "new-session-id", nil)

	req := httptest.NewRequest(http.MethodGet, "/callback?code=valid_code", nil)
	rr := httptest.NewRecorder()
	httpHandler := handler.VkOAuthCallback(conf, redirectTo, mockVkOAuth)
	httpHandler(rr, req)

	assert.Equal(t, http.StatusFound, rr.Code)
	assert.Equal(t, redirectTo, rr.Header().Get("Location"))
	mockVkOAuth.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}
