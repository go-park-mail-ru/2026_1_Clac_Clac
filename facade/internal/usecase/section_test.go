package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/domain"
	mockSectionClient "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/usecase/mock_section_client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	fixedSectionLink = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	sectionBoardLink = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	errSectionTest   = errors.New("client error")
)

func TestSectionGetSections(t *testing.T) {
	expected := []domain.SectionInfo{
		{Link: fixedSectionLink, Name: "To Do"},
	}

	tests := []struct {
		name         string
		mockBehavior func(m *mockSectionClient.SectionClient)
		expected     []domain.SectionInfo
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockSectionClient.SectionClient) {
				m.On("GetSections", context.Background(), domain.GetSectionsRequest{
					UserLink:  fixedUserLink,
					BoardLink: sectionBoardLink,
				}).Return(expected, nil)
			},
			expected:    expected,
			expectError: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockSectionClient.SectionClient) {
				m.On("GetSections", context.Background(), domain.GetSectionsRequest{
					UserLink:  fixedUserLink,
					BoardLink: sectionBoardLink,
				}).Return([]domain.SectionInfo(nil), errSectionTest)
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockSectionClient.NewSectionClient(t)
			tc.mockBehavior(m)

			result, err := NewSection(m).GetSections(context.Background(), domain.GetSectionsRequest{
				UserLink:  fixedUserLink,
				BoardLink: sectionBoardLink,
			})

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestSectionGetSection(t *testing.T) {
	req := domain.GetSectionRequest{UserLink: fixedUserLink, SectionLink: fixedSectionLink}
	expected := domain.SectionInfo{Link: fixedSectionLink, Name: "To Do"}

	tests := []struct {
		name         string
		mockBehavior func(m *mockSectionClient.SectionClient)
		expected     domain.SectionInfo
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockSectionClient.SectionClient) {
				m.On("GetSection", context.Background(), req).Return(expected, nil)
			},
			expected:    expected,
			expectError: false,
		},
		{
			name: "NotFound",
			mockBehavior: func(m *mockSectionClient.SectionClient) {
				m.On("GetSection", context.Background(), req).Return(domain.SectionInfo{}, common.ErrorNonexistentUser)
			},
			expected:    domain.SectionInfo{},
			expectError: true,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockSectionClient.SectionClient) {
				m.On("GetSection", context.Background(), req).Return(domain.SectionInfo{}, errSectionTest)
			},
			expected:    domain.SectionInfo{},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockSectionClient.NewSectionClient(t)
			tc.mockBehavior(m)

			result, err := NewSection(m).GetSection(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestSectionGetCards(t *testing.T) {
	req := domain.GetCardsRequest{UserLink: fixedUserLink, SectionLink: fixedSectionLink}
	expected := []domain.CardInfo{
		{CardLink: uuid.New(), Title: "Task 1"},
	}

	tests := []struct {
		name         string
		mockBehavior func(m *mockSectionClient.SectionClient)
		expected     []domain.CardInfo
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockSectionClient.SectionClient) {
				m.On("GetCards", context.Background(), req).Return(expected, nil)
			},
			expected:    expected,
			expectError: false,
		},
		{
			name: "NotFound",
			mockBehavior: func(m *mockSectionClient.SectionClient) {
				m.On("GetCards", context.Background(), req).Return([]domain.CardInfo(nil), common.ErrorNonexistentUser)
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockSectionClient.SectionClient) {
				m.On("GetCards", context.Background(), req).Return([]domain.CardInfo(nil), errSectionTest)
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockSectionClient.NewSectionClient(t)
			tc.mockBehavior(m)

			result, err := NewSection(m).GetCards(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestSectionCreateSection(t *testing.T) {
	req := domain.CreateSectionRequest{
		UserLink:    fixedUserLink,
		BoardLink:   sectionBoardLink,
		Name:        "New Section",
		IsMandatory: false,
		Color:       "red",
	}
	expected := domain.SectionInfo{Link: fixedSectionLink, Name: "New Section"}

	tests := []struct {
		name         string
		mockBehavior func(m *mockSectionClient.SectionClient)
		expected     domain.SectionInfo
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockSectionClient.SectionClient) {
				m.On("CreateSection", context.Background(), req).Return(expected, nil)
			},
			expected:    expected,
			expectError: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockSectionClient.SectionClient) {
				m.On("CreateSection", context.Background(), req).Return(domain.SectionInfo{}, errSectionTest)
			},
			expected:    domain.SectionInfo{},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockSectionClient.NewSectionClient(t)
			tc.mockBehavior(m)

			result, err := NewSection(m).CreateSection(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestSectionDeleteSection(t *testing.T) {
	req := domain.DeleteSectionRequest{UserLink: fixedUserLink, SectionLink: fixedSectionLink}

	tests := []struct {
		name         string
		mockBehavior func(m *mockSectionClient.SectionClient)
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockSectionClient.SectionClient) {
				m.On("DeleteSection", context.Background(), req).Return(nil)
			},
			expectError: false,
		},
		{
			name: "NotFound",
			mockBehavior: func(m *mockSectionClient.SectionClient) {
				m.On("DeleteSection", context.Background(), req).Return(common.ErrorNonexistentUser)
			},
			expectError: true,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockSectionClient.SectionClient) {
				m.On("DeleteSection", context.Background(), req).Return(errSectionTest)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockSectionClient.NewSectionClient(t)
			tc.mockBehavior(m)

			err := NewSection(m).DeleteSection(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSectionReorderSection(t *testing.T) {
	req := domain.ReorderSectionRequest{
		UserLink:  fixedUserLink,
		BoardLink: sectionBoardLink,
		LinksList: []uuid.UUID{fixedSectionLink},
	}

	tests := []struct {
		name         string
		mockBehavior func(m *mockSectionClient.SectionClient)
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockSectionClient.SectionClient) {
				m.On("ReorderSection", context.Background(), req).Return(nil)
			},
			expectError: false,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockSectionClient.SectionClient) {
				m.On("ReorderSection", context.Background(), req).Return(errSectionTest)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockSectionClient.NewSectionClient(t)
			tc.mockBehavior(m)

			err := NewSection(m).ReorderSection(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSectionUpdateSection(t *testing.T) {
	req := domain.UpdateSectionRequest{
		UserLink:    fixedUserLink,
		SectionLink: fixedSectionLink,
		Name:        "Updated",
	}

	tests := []struct {
		name         string
		mockBehavior func(m *mockSectionClient.SectionClient)
		expectError  bool
	}{
		{
			name: "Success",
			mockBehavior: func(m *mockSectionClient.SectionClient) {
				m.On("UpdateSection", context.Background(), req).Return(nil)
			},
			expectError: false,
		},
		{
			name: "NotFound",
			mockBehavior: func(m *mockSectionClient.SectionClient) {
				m.On("UpdateSection", context.Background(), req).Return(common.ErrorNonexistentUser)
			},
			expectError: true,
		},
		{
			name: "ClientError",
			mockBehavior: func(m *mockSectionClient.SectionClient) {
				m.On("UpdateSection", context.Background(), req).Return(errSectionTest)
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mockSectionClient.NewSectionClient(t)
			tc.mockBehavior(m)

			err := NewSection(m).UpdateSection(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
