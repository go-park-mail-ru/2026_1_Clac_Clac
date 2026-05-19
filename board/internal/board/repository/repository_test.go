package repository_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/repository"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/config"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/boardRbac"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/s3"
)

type MockS3Bucket struct {
	mock.Mock
}

func (m *MockS3Bucket) Put(ctx context.Context, data io.Reader, key string, contentType string) (string, error) {
	args := m.Called(ctx, data, key, contentType)
	return args.String(0), args.Error(1)
}

func (m *MockS3Bucket) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) NewBucket(bucket string, prefix string, action s3.Action) s3.S3Bucket {
	args := m.Called(bucket, prefix, action)
	return args.Get(0).(s3.S3Bucket)
}

func setupRepo(dbMock pgxmock.PgxPoolIface, s3BucketMock s3.S3Bucket) *repository.Repository {
	repoConf := repository.Config{
		CreateBoardDefaultUserRole: config.DefaultBoardConfig().Repository.CreateBoardDefaultUserRole,
	}
	return repository.NewRepository(dbMock, s3BucketMock, repoConf)
}

func TestGetBoards(t *testing.T) {
	userID1 := uuid.New()

	board1 := dto.BoardEntry{Link: uuid.New(), Name: "board 1", CreatedAt: time.Now()}
	board2 := dto.BoardEntry{Link: uuid.New(), Name: "board 2", CreatedAt: time.Now()}

	tests := []struct {
		Name           string
		TargetId       uuid.UUID
		MockSetup      func(dbMock pgxmock.PgxPoolIface, targetID uuid.UUID)
		ExpectedBoards []dto.BoardEntry
	}{
		{
			Name:     "success get user boards",
			TargetId: userID1,
			MockSetup: func(dbMock pgxmock.PgxPoolIface, targetID uuid.UUID) {
				rows := pgxmock.NewRows([]string{"link", "name", "description", "background", "created_at"}).
					AddRow(board1.Link, board1.Name, board1.Description, board1.Background, board1.CreatedAt).
					AddRow(board2.Link, board2.Name, board2.Description, board2.Background, board2.CreatedAt)

				query := `(?s)SELECT b.link, b.name, b.description, b.background, b.created_at.*`
				dbMock.ExpectQuery(query).
					WithArgs(targetID).
					WillReturnRows(rows)
			},
			ExpectedBoards: []dto.BoardEntry{board1, board2},
		},
		{
			Name:     "user has no boards",
			TargetId: userID1,
			MockSetup: func(dbMock pgxmock.PgxPoolIface, targetID uuid.UUID) {
				rows := pgxmock.NewRows([]string{"link", "name", "description", "background", "created_at"})
				query := `(?s)SELECT b.link, b.name, b.description, b.background, b.created_at.*`
				dbMock.ExpectQuery(query).
					WithArgs(targetID).
					WillReturnRows(rows)
			},
			ExpectedBoards: []dto.BoardEntry{},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			dbMock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer dbMock.Close()

			test.MockSetup(dbMock, test.TargetId)

			repoBoards := setupRepo(dbMock, new(MockS3Bucket))
			ctx := context.Background()

			boards, err := repoBoards.GetBoards(ctx, test.TargetId)

			assert.NoError(t, err)
			assert.Equal(t, test.ExpectedBoards, boards)

			assert.NoError(t, dbMock.ExpectationsWereMet())
		})
	}
}

func TestGetBoard(t *testing.T) {
	boardLink := uuid.New()
	now := time.Now()

	expectedBoard := dto.BoardEntry{
		Link:        boardLink,
		Name:        "Single Board",
		Description: "Desc",
		Background:  "#fff",
		CreatedAt:   now,
	}

	tests := []struct {
		Name          string
		BoardLink     uuid.UUID
		MockSetup     func(dbMock pgxmock.PgxPoolIface)
		ExpectedBoard dto.BoardEntry
		ExpectedErr   error
	}{
		{
			Name:      "success get board",
			BoardLink: boardLink,
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"link", "name", "description", "background", "created_at"}).
					AddRow(expectedBoard.Link, expectedBoard.Name, expectedBoard.Description, expectedBoard.Background, expectedBoard.CreatedAt)

				query := `(?s)SELECT link, name, description, background, created_at FROM board_actual.*`
				dbMock.ExpectQuery(query).
					WithArgs(boardLink).
					WillReturnRows(rows)
			},
			ExpectedBoard: expectedBoard,
			ExpectedErr:   nil,
		},
		{
			Name:      "board not found",
			BoardLink: boardLink,
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				query := `(?s)SELECT link, name, description, background, created_at FROM board_actual.*`
				dbMock.ExpectQuery(query).
					WithArgs(boardLink).
					WillReturnError(pgx.ErrNoRows)
			},
			ExpectedBoard: dto.BoardEntry{},
			ExpectedErr:   common.ErrBoardNotFound,
		},
		{
			Name:      "db error",
			BoardLink: boardLink,
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				query := `(?s)SELECT link, name, description, background, created_at FROM board_actual.*`
				dbMock.ExpectQuery(query).
					WithArgs(boardLink).
					WillReturnError(fmt.Errorf("db error"))
			},
			ExpectedBoard: dto.BoardEntry{},
			ExpectedErr:   fmt.Errorf("pool.Query: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			dbMock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer dbMock.Close()

			tt.MockSetup(dbMock)

			repo := setupRepo(dbMock, new(MockS3Bucket))
			board, err := repo.GetBoard(context.Background(), tt.BoardLink)

			if tt.ExpectedErr != nil {
				assert.Error(t, err)
				if tt.ExpectedErr == common.ErrBoardNotFound {
					assert.ErrorIs(t, err, common.ErrBoardNotFound)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.ExpectedBoard, board)
			}
			assert.NoError(t, dbMock.ExpectationsWereMet())
		})
	}
}

func TestCreateBoard(t *testing.T) {
	authorID := uuid.New()
	newBoardLink := uuid.New()
	mockDBErr := errors.New("db error")
	now := time.Now().Truncate(time.Second)

	boardInfo := dto.NewBoardInfo{
		Name:        "Nexus Core",
		Description: "Main board",
		Background:  "#1e1e2e",
	}

	expectedEntry := dto.BoardEntry{
		Link:        newBoardLink,
		Name:        boardInfo.Name,
		Description: boardInfo.Description,
		Background:  boardInfo.Background,
		CreatedAt:   now,
	}

	tests := []struct {
		Name          string
		BoardInfo     dto.NewBoardInfo
		AuthorLink    uuid.UUID
		MockSetup     func(dbMock pgxmock.PgxPoolIface)
		ExpectedEntry dto.BoardEntry
		ExpectedErr   error
	}{
		{
			Name:       "success create board",
			BoardInfo:  boardInfo,
			AuthorLink: authorID,
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				dbMock.ExpectBegin()

				rows := pgxmock.NewRows([]string{"board_id", "link", "created_at"}).
					AddRow(1, newBoardLink, now)
				dbMock.ExpectQuery(`(?s)INSERT INTO board DEFAULT VALUES.*`).
					WillReturnRows(rows)

				dbMock.ExpectExec(`(?s)INSERT INTO board_version.*`).
					WithArgs(1, boardInfo.Name, boardInfo.Description, boardInfo.Background).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				dbMock.ExpectExec(`(?s)INSERT INTO member_board.*`).
					WithArgs(newBoardLink, authorID, config.DefaultBoardConfig().Repository.CreateBoardDefaultUserRole).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				dbMock.ExpectCommit()
			},
			ExpectedEntry: expectedEntry,
			ExpectedErr:   nil,
		},
		{
			Name:       "error on create board version (NotNullViolation)",
			BoardInfo:  boardInfo,
			AuthorLink: authorID,
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				dbMock.ExpectBegin()

				rows := pgxmock.NewRows([]string{"board_id", "link", "created_at"}).AddRow(1, newBoardLink, now)
				dbMock.ExpectQuery(`(?s)INSERT INTO board DEFAULT VALUES.*`).WillReturnRows(rows)

				dbMock.ExpectExec(`(?s)INSERT INTO board_version.*`).
					WithArgs(1, boardInfo.Name, boardInfo.Description, boardInfo.Background).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.NotNullViolation})

				dbMock.ExpectRollback()
			},
			ExpectedEntry: dto.BoardEntry{},
			ExpectedErr:   common.ErrNotNullValue,
		},
		{
			Name:       "error on create board version (CheckViolation)",
			BoardInfo:  boardInfo,
			AuthorLink: authorID,
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				dbMock.ExpectBegin()

				rows := pgxmock.NewRows([]string{"board_id", "link", "created_at"}).AddRow(1, newBoardLink, now)
				dbMock.ExpectQuery(`(?s)INSERT INTO board DEFAULT VALUES.*`).WillReturnRows(rows)

				dbMock.ExpectExec(`(?s)INSERT INTO board_version.*`).
					WithArgs(1, boardInfo.Name, boardInfo.Description, boardInfo.Background).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.CheckViolation})

				dbMock.ExpectRollback()
			},
			ExpectedEntry: dto.BoardEntry{},
			ExpectedErr:   common.ErrInvalidBoardData,
		},
		{
			Name:       "error on create board member (UniqueViolation)",
			BoardInfo:  boardInfo,
			AuthorLink: authorID,
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				dbMock.ExpectBegin()

				rows := pgxmock.NewRows([]string{"board_id", "link", "created_at"}).AddRow(1, newBoardLink, now)
				dbMock.ExpectQuery(`(?s)INSERT INTO board DEFAULT VALUES.*`).WillReturnRows(rows)

				dbMock.ExpectExec(`(?s)INSERT INTO board_version.*`).
					WithArgs(1, boardInfo.Name, boardInfo.Description, boardInfo.Background).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				dbMock.ExpectExec(`(?s)INSERT INTO member_board.*`).
					WithArgs(newBoardLink, authorID, config.DefaultBoardConfig().Repository.CreateBoardDefaultUserRole).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

				dbMock.ExpectRollback()
			},
			ExpectedEntry: dto.BoardEntry{},
			ExpectedErr:   common.ErrUserAlreadyMember,
		},
		{
			Name:       "error on create board version (Generic DB error)",
			BoardInfo:  boardInfo,
			AuthorLink: authorID,
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				dbMock.ExpectBegin()

				rows := pgxmock.NewRows([]string{"board_id", "link", "created_at"}).AddRow(1, newBoardLink, now)
				dbMock.ExpectQuery(`(?s)INSERT INTO board DEFAULT VALUES.*`).WillReturnRows(rows)

				dbMock.ExpectExec(`(?s)INSERT INTO board_version.*`).
					WithArgs(1, boardInfo.Name, boardInfo.Description, boardInfo.Background).
					WillReturnError(mockDBErr)

				dbMock.ExpectRollback()
			},
			ExpectedEntry: dto.BoardEntry{},
			ExpectedErr:   mockDBErr,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			dbMock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer dbMock.Close()

			test.MockSetup(dbMock)

			repo := setupRepo(dbMock, new(MockS3Bucket))
			ctx := context.Background()

			entry, err := repo.CreateBoard(ctx, test.BoardInfo, test.AuthorLink)

			if test.ExpectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, test.ExpectedErr)
				assert.Equal(t, dto.BoardEntry{}, entry)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.ExpectedEntry, entry)
			}

			assert.NoError(t, dbMock.ExpectationsWereMet())
		})
	}
}

func TestDeleteBoard(t *testing.T) {
	boardLink := uuid.New()

	tests := []struct {
		Name        string
		BoardLink   uuid.UUID
		MockSetup   func(dbMock pgxmock.PgxPoolIface)
		ExpectedErr error
	}{
		{
			Name:      "success delete board",
			BoardLink: boardLink,
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				dbMock.ExpectExec(`(?s)DELETE FROM board.*`).
					WithArgs(boardLink).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			ExpectedErr: nil,
		},
		{
			Name:      "board not found (0 rows affected)",
			BoardLink: boardLink,
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				dbMock.ExpectExec(`(?s)DELETE FROM board.*`).
					WithArgs(boardLink).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			ExpectedErr: common.ErrBoardNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			dbMock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer dbMock.Close()

			test.MockSetup(dbMock)

			repo := setupRepo(dbMock, new(MockS3Bucket))
			err = repo.DeleteBoard(context.Background(), test.BoardLink)

			if test.ExpectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, dbMock.ExpectationsWereMet())
		})
	}
}

func TestUpdateBoard(t *testing.T) {
	boardLink := uuid.New()
	boardInfo := dto.UpdateBoardInfo{
		Link:        boardLink,
		Name:        "Nexus Core Updated",
		Description: "Updated main board",
		Background:  "#2e2e3e",
	}

	tests := []struct {
		Name        string
		BoardInfo   dto.UpdateBoardInfo
		MockSetup   func(dbMock pgxmock.PgxPoolIface)
		ExpectedErr error
	}{
		{
			Name:      "success update board",
			BoardInfo: boardInfo,
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				dbMock.ExpectExec(`(?s)UPDATE board_actual.*`).
					WithArgs(boardInfo.Name, boardInfo.Description, boardInfo.Background, boardInfo.Link).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			ExpectedErr: nil,
		},
		{
			Name:      "board not found (0 rows affected)",
			BoardInfo: boardInfo,
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				dbMock.ExpectExec(`(?s)UPDATE board_actual.*`).
					WithArgs(boardInfo.Name, boardInfo.Description, boardInfo.Background, boardInfo.Link).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			ExpectedErr: common.ErrBoardNotFound,
		},
		{
			Name:      "error update board missing field (NotNullViolation)",
			BoardInfo: boardInfo,
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				dbMock.ExpectExec(`(?s)UPDATE board_actual.*`).
					WithArgs(boardInfo.Name, boardInfo.Description, boardInfo.Background, boardInfo.Link).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.NotNullViolation})
			},
			ExpectedErr: common.ErrNotNullValue,
		},
		{
			Name:      "error update board invalid data (CheckViolation)",
			BoardInfo: boardInfo,
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				dbMock.ExpectExec(`(?s)UPDATE board_actual.*`).
					WithArgs(boardInfo.Name, boardInfo.Description, boardInfo.Background, boardInfo.Link).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.CheckViolation})
			},
			ExpectedErr: common.ErrInvalidBoardData,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			dbMock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer dbMock.Close()

			test.MockSetup(dbMock)

			repo := setupRepo(dbMock, new(MockS3Bucket))
			err = repo.UpdateBoard(context.Background(), test.BoardInfo)

			if test.ExpectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, test.ExpectedErr)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, dbMock.ExpectationsWereMet())
		})
	}
}


func TestUploadBackground(t *testing.T) {
	tests := []struct {
		Name        string
		Filename    string
		ContentType string
		SetupMock   func(s3Mock *MockS3Bucket, reader io.Reader)
		ExpectedKey string
		ExpectError bool
	}{
		{
			Name:        "success upload",
			Filename:    "bg.jpg",
			ContentType: "image/jpeg",
			SetupMock: func(s3Mock *MockS3Bucket, reader io.Reader) {
				s3Mock.On("Put", mock.Anything, reader, "bg.jpg", "image/jpeg").
					Return("https://s3.local/bg.jpg", nil)
			},
			ExpectedKey: "https://s3.local/bg.jpg",
			ExpectError: false,
		},
		{
			Name:        "s3 error",
			Filename:    "fail.jpg",
			ContentType: "image/jpeg",
			SetupMock: func(s3Mock *MockS3Bucket, reader io.Reader) {
				s3Mock.On("Put", mock.Anything, reader, "fail.jpg", "image/jpeg").
					Return("", fmt.Errorf("s3 timeout"))
			},
			ExpectedKey: "",
			ExpectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			dbMock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer dbMock.Close()

			s3Mock := new(MockS3Bucket)
			repo := setupRepo(dbMock, s3Mock)

			reader := bytes.NewReader([]byte("fake image data"))
			tt.SetupMock(s3Mock, reader)

			key, err := repo.UploadBackground(context.Background(), reader, tt.Filename, tt.ContentType)

			if tt.ExpectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.ExpectedKey, key)
			}

			s3Mock.AssertExpectations(t)
		})
	}
}

func TestUpdateBackground(t *testing.T) {
	boardLink := uuid.New()
	newBackground := "https://s3.local/new_bg.jpg"

	tests := []struct {
		Name        string
		MockSetup   func(dbMock pgxmock.PgxPoolIface)
		ExpectedErr error
	}{
		{
			Name: "success update background",
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				query := `UPDATE board_actual SET background = $1 WHERE link = $2`
				dbMock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(newBackground, boardLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			ExpectedErr: nil,
		},
		{
			Name: "board not found (0 rows affected)",
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				query := `UPDATE board_actual SET background = $1 WHERE link = $2`
				dbMock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(newBackground, boardLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			ExpectedErr: common.ErrBoardNotFound,
		},
		{
			Name: "db error",
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				query := `UPDATE board_actual SET background = $1 WHERE link = $2`
				dbMock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(newBackground, boardLink).
					WillReturnError(fmt.Errorf("db connection dropped"))
			},
			ExpectedErr: fmt.Errorf("update board: db connection dropped"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			dbMock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer dbMock.Close()

			tt.MockSetup(dbMock)
			repo := setupRepo(dbMock, new(MockS3Bucket))

			err = repo.UpdateBackground(context.Background(), newBackground, boardLink)

			if tt.ExpectedErr != nil {
				assert.Error(t, err)
				if tt.ExpectedErr == common.ErrBoardNotFound {
					assert.ErrorIs(t, err, common.ErrBoardNotFound)
				}
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, dbMock.ExpectationsWereMet())
		})
	}
}

func TestGetUsersOfBoard(t *testing.T) {
	boardLink := uuid.New()
	user1 := uuid.New()
	user2 := uuid.New()

	tests := []struct {
		Name        string
		BoardLink   uuid.UUID
		MockSetup   func(dbMock pgxmock.PgxPoolIface)
		Expected    []uuid.UUID
		ExpectedErr error
	}{
		{
			Name:      "success get users of board",
			BoardLink: boardLink,
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"user_link"}).
					AddRow(user1).
					AddRow(user2)

				dbMock.ExpectQuery(`(?s)SELECT user_link FROM member_board.*`).
					WithArgs(boardLink).
					WillReturnRows(rows)
			},
			Expected:    []uuid.UUID{user1, user2},
			ExpectedErr: nil,
		},
		{
			Name:      "board not found (empty list)",
			BoardLink: boardLink,
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"user_link"})

				dbMock.ExpectQuery(`(?s)SELECT user_link FROM member_board.*`).
					WithArgs(boardLink).
					WillReturnRows(rows)
			},
			Expected:    []uuid.UUID{},
			ExpectedErr: common.ErrBoardNotFound,
		},
		{
			Name:      "db query error",
			BoardLink: boardLink,
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				dbMock.ExpectQuery(`(?s)SELECT user_link FROM member_board.*`).
					WithArgs(boardLink).
					WillReturnError(fmt.Errorf("db connection dropped"))
			},
			Expected:    []uuid.UUID{},
			ExpectedErr: fmt.Errorf("pool.Query: db connection dropped"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			dbMock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer dbMock.Close()

			tt.MockSetup(dbMock)

			repo := setupRepo(dbMock, new(MockS3Bucket))
			users, err := repo.GetUsersOfBoard(context.Background(), tt.BoardLink)

			if tt.ExpectedErr != nil {
				assert.Error(t, err)
				if tt.ExpectedErr == common.ErrBoardNotFound {
					assert.ErrorIs(t, err, common.ErrBoardNotFound)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.Expected, users)
			}
			assert.NoError(t, dbMock.ExpectationsWereMet())
		})
	}
}

func TestCreateInviteRepo(t *testing.T) {
	boardLink := uuid.New()
	now := time.Now()

	inviteInfo := dto.NewInviteInfo{
		BoardLink:   boardLink,
		UserLink:    nil,
		DefaultRole: rbac.Roles.Editor,
		ExpireTime:  &now,
	}

	tests := []struct {
		Name        string
		MockSetup   func(dbMock pgxmock.PgxPoolIface)
		ExpectedErr error
	}{
		{
			Name: "foreign key violation",
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				dbMock.ExpectQuery(`(?s)INSERT INTO invite.*`).
					WithArgs(boardLink, (*uuid.UUID)(nil), rbac.Roles.Editor, &now).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})
			},
			ExpectedErr: common.ErrInvalidBoardReference,
		},
		{
			Name: "not null violation",
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				dbMock.ExpectQuery(`(?s)INSERT INTO invite.*`).
					WithArgs(boardLink, (*uuid.UUID)(nil), rbac.Roles.Editor, &now).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.NotNullViolation})
			},
			ExpectedErr: common.ErrNotNullValue,
		},
		{
			Name: "db error",
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				dbMock.ExpectQuery(`(?s)INSERT INTO invite.*`).
					WithArgs(boardLink, (*uuid.UUID)(nil), rbac.Roles.Editor, &now).
					WillReturnError(fmt.Errorf("connection timeout"))
			},
			ExpectedErr: fmt.Errorf("connection timeout"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			dbMock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer dbMock.Close()

			tt.MockSetup(dbMock)

			repo := setupRepo(dbMock, new(MockS3Bucket))
			_, err = repo.CreateInvite(context.Background(), inviteInfo)

			if tt.ExpectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, dbMock.ExpectationsWereMet())
		})
	}
}

func TestGetInviteByLinkRepo(t *testing.T) {
	inviteLink := uuid.New()

	tests := []struct {
		Name        string
		MockSetup   func(dbMock pgxmock.PgxPoolIface)
		ExpectedErr error
	}{
		{
			Name: "not found",
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				dbMock.ExpectQuery(`(?s)SELECT.*FROM invite.*`).
					WithArgs(inviteLink).
					WillReturnError(pgx.ErrNoRows)
			},
			ExpectedErr: common.ErrInviteNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			dbMock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer dbMock.Close()

			tt.MockSetup(dbMock)

			repo := setupRepo(dbMock, new(MockS3Bucket))
			_, err = repo.GetInviteByLink(context.Background(), inviteLink)

			assert.Error(t, err)
			assert.ErrorIs(t, err, tt.ExpectedErr)

			assert.NoError(t, dbMock.ExpectationsWereMet())
		})
	}
}

func TestAddMemberToBoardRepo(t *testing.T) {
	boardLink := uuid.New()
	userLink := uuid.New()

	tests := []struct {
		Name        string
		MockSetup   func(dbMock pgxmock.PgxPoolIface)
		ExpectedErr error
	}{
		{
			Name: "success add member",
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				dbMock.ExpectExec(`(?s)INSERT INTO member_board.*`).
					WithArgs(boardLink, userLink, rbac.Roles.Editor).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
		},
		{
			Name: "unique violation",
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				dbMock.ExpectExec(`(?s)INSERT INTO member_board.*`).
					WithArgs(boardLink, userLink, rbac.Roles.Editor).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
			},
			ExpectedErr: common.ErrUserAlreadyMember,
		},
		{
			Name: "foreign key violation",
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				dbMock.ExpectExec(`(?s)INSERT INTO member_board.*`).
					WithArgs(boardLink, userLink, rbac.Roles.Editor).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.ForeignKeyViolation})
			},
			ExpectedErr: common.ErrInvalidBoardReference,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			dbMock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer dbMock.Close()

			tt.MockSetup(dbMock)

			repo := setupRepo(dbMock, new(MockS3Bucket))
			err = repo.AddMemberToBoard(context.Background(), boardLink, userLink, rbac.Roles.Editor)

			if tt.ExpectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, dbMock.ExpectationsWereMet())
		})
	}
}

func TestCloseInviteRepo(t *testing.T) {
	inviteLink := uuid.New()

	tests := []struct {
		Name        string
		MockSetup   func(dbMock pgxmock.PgxPoolIface)
		ExpectedErr error
	}{
		{
			Name: "success close invite",
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				dbMock.ExpectExec(`(?s)UPDATE invite SET status.*`).
					WithArgs(inviteLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
		},
		{
			Name: "not found (no rows affected)",
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				dbMock.ExpectExec(`(?s)UPDATE invite SET status.*`).
					WithArgs(inviteLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			ExpectedErr: common.ErrInviteNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			dbMock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer dbMock.Close()

			tt.MockSetup(dbMock)

			repo := setupRepo(dbMock, new(MockS3Bucket))
			err = repo.CloseInvite(context.Background(), inviteLink)

			if tt.ExpectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, dbMock.ExpectationsWereMet())
		})
	}
}
