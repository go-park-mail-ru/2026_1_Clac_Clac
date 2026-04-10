package repository_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/repository"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/s3"
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
	s3ClientMock := new(MockS3Client)
	conf := &config.S3{
		BoardsBackgroundsBucket: "test-bucket",
		BoardsBackgroundsPrefix: "test-prefix",
	}

	s3ClientMock.On("NewBucket", conf.BoardsBackgroundsBucket, conf.BoardsBackgroundsPrefix, s3.ACL.PublicRead).Return(s3BucketMock)

	return repository.NewRepository(dbMock, s3ClientMock, *conf, config.DefaultBoardConfig().Repository)
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
				getBoardQuery := `
					SELECT b.link, b.name, b.description, b.background, b.created_at
					FROM board_actual b
					JOIN member_board mb ON b.link = mb.board_link
					WHERE mb.user_link = $1
				`
				rows := pgxmock.NewRows([]string{"link", "name", "description", "background", "created_at"}).
					AddRow(board1.Link, board1.Name, board1.Description, board1.Background, board1.CreatedAt).
					AddRow(board2.Link, board2.Name, board2.Description, board2.Background, board2.CreatedAt)

				dbMock.ExpectQuery(regexp.QuoteMeta(getBoardQuery)).
					WithArgs(targetID).
					WillReturnRows(rows)
			},
			ExpectedBoards: []dto.BoardEntry{board1, board2},
		},
		{
			Name:     "user has no boards",
			TargetId: userID1,
			MockSetup: func(dbMock pgxmock.PgxPoolIface, targetID uuid.UUID) {
				getBoardQuery := `
					SELECT b.link, b.name, b.description, b.background, b.created_at
					FROM board_actual b
					JOIN member_board mb ON b.link = mb.board_link
					WHERE mb.user_link = $1
				`
				rows := pgxmock.NewRows([]string{"link", "name", "description", "background", "created_at"})

				dbMock.ExpectQuery(regexp.QuoteMeta(getBoardQuery)).
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
				query := `
					SELECT link, name, description, background, created_at
					FROM board_actual
					WHERE link = $1
				`
				rows := pgxmock.NewRows([]string{"link", "name", "description", "background", "created_at"}).
					AddRow(expectedBoard.Link, expectedBoard.Name, expectedBoard.Description, expectedBoard.Background, expectedBoard.CreatedAt)

				dbMock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(boardLink).WillReturnRows(rows)
			},
			ExpectedBoard: expectedBoard,
			ExpectedErr:   nil,
		},
		{
			Name:      "board not found",
			BoardLink: boardLink,
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				query := `
					SELECT link, name, description, background, created_at
					FROM board_actual
					WHERE link = $1
				`
				dbMock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(boardLink).WillReturnError(pgx.ErrNoRows)
			},
			ExpectedBoard: dto.BoardEntry{},
			ExpectedErr:   common.ErrBoardNotFound,
		},
		{
			Name:      "db error",
			BoardLink: boardLink,
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				query := `
					SELECT link, name, description, background, created_at
					FROM board_actual
					WHERE link = $1
				`
				dbMock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(boardLink).WillReturnError(fmt.Errorf("db error"))
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
		ExpectError   bool
	}{
		{
			Name:       "success create board",
			BoardInfo:  boardInfo,
			AuthorLink: authorID,
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				dbMock.ExpectBegin()

				createBoardQuery := `INSERT INTO board DEFAULT VALUES RETURNING board_id, link, created_at`
				rows := pgxmock.NewRows([]string{"board_id", "link", "created_at"}).
					AddRow(1, newBoardLink, now)
				dbMock.ExpectQuery(regexp.QuoteMeta(createBoardQuery)).
					WillReturnRows(rows)

				createVersionQuery := `INSERT INTO board_version (board_id, board_name, description_board, url_path_background) VALUES ($1, $2, $3, $4)`
				dbMock.ExpectExec(regexp.QuoteMeta(createVersionQuery)).
					WithArgs(1, boardInfo.Name, boardInfo.Description, boardInfo.Background).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				createMemberQuery := `INSERT INTO member_board (board_link, user_link, level_member) VALUES ($1, $2, $3::user_level)`
				dbMock.ExpectExec(regexp.QuoteMeta(createMemberQuery)).
					WithArgs(newBoardLink, authorID, "creator").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				dbMock.ExpectCommit()
			},
			ExpectedEntry: expectedEntry,
			ExpectError:   false,
		},
		{
			Name:       "error on create board version",
			BoardInfo:  boardInfo,
			AuthorLink: authorID,
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				dbMock.ExpectBegin()

				createBoardQuery := `INSERT INTO board DEFAULT VALUES RETURNING board_id, link, created_at`
				rows := pgxmock.NewRows([]string{"board_id", "link", "created_at"}).
					AddRow(1, newBoardLink, now)
				dbMock.ExpectQuery(regexp.QuoteMeta(createBoardQuery)).
					WillReturnRows(rows)

				createVersionQuery := `INSERT INTO board_version (board_id, board_name, description_board, url_path_background) VALUES ($1, $2, $3, $4)`
				dbMock.ExpectExec(regexp.QuoteMeta(createVersionQuery)).
					WithArgs(1, boardInfo.Name, boardInfo.Description, boardInfo.Background).
					WillReturnError(fmt.Errorf("db error"))

				dbMock.ExpectRollback()
			},
			ExpectedEntry: dto.BoardEntry{},
			ExpectError:   true,
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

			if test.ExpectError {
				assert.Error(t, err)
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
				deleteBoardQuery := `DELETE FROM board WHERE board.link = $1`

				dbMock.ExpectExec(regexp.QuoteMeta(deleteBoardQuery)).
					WithArgs(boardLink).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			ExpectedErr: nil,
		},
		{
			Name:      "board not found (0 rows affected)",
			BoardLink: boardLink,
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				deleteBoardQuery := `DELETE FROM board WHERE board.link = $1`

				dbMock.ExpectExec(regexp.QuoteMeta(deleteBoardQuery)).
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
				updateBoardQuery := `
					UPDATE board_actual
					SET name = $1, description = $2, background = $3
					WHERE link = $4
				`

				dbMock.ExpectExec(regexp.QuoteMeta(updateBoardQuery)).
					WithArgs(boardInfo.Name, boardInfo.Description, boardInfo.Background, boardInfo.Link).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			ExpectedErr: nil,
		},
		{
			Name:      "board not found (0 rows affected)",
			BoardInfo: boardInfo,
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				updateBoardQuery := `
					UPDATE board_actual
					SET name = $1, description = $2, background = $3
					WHERE link = $4
				`

				dbMock.ExpectExec(regexp.QuoteMeta(updateBoardQuery)).
					WithArgs(boardInfo.Name, boardInfo.Description, boardInfo.Background, boardInfo.Link).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
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
			err = repo.UpdateBoard(context.Background(), test.BoardInfo)

			if test.ExpectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, dbMock.ExpectationsWereMet())
		})
	}
}

func TestGetUserRoleOnBoard(t *testing.T) {
	userLink := uuid.New()
	boardLink := uuid.New()

	tests := []struct {
		Name         string
		MockSetup    func(dbMock pgxmock.PgxPoolIface)
		ExpectedRole common.Role
		ExpectError  bool
	}{
		{
			Name: "success get role",
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				query := `
					SELECT level_member FROM member_board
					WHERE board_link = $1 AND user_link = $2;
				`
				rows := pgxmock.NewRows([]string{"level_member"}).AddRow(common.Role("creator"))
				dbMock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(boardLink, userLink).
					WillReturnRows(rows)
			},
			ExpectedRole: common.Role("creator"),
			ExpectError:  false,
		},
		{
			Name: "role not found (no rows)",
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				query := `
					SELECT level_member FROM member_board
					WHERE board_link = $1 AND user_link = $2;
				`
				dbMock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(boardLink, userLink).
					WillReturnError(pgx.ErrNoRows)
			},
			ExpectedRole: common.Roles.None,
			ExpectError:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			dbMock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer dbMock.Close()

			test.MockSetup(dbMock)

			repo := setupRepo(dbMock, new(MockS3Bucket))
			role, err := repo.GetUserRoleOnBoard(context.Background(), userLink, boardLink)

			if test.ExpectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.ExpectedRole, role)
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
				query := `
					UPDATE board_actual
					SET background = $1
					WHERE link = $2
				`
				dbMock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(newBackground, boardLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			ExpectedErr: nil,
		},
		{
			Name: "board not found (0 rows affected)",
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				query := `
					UPDATE board_actual
					SET background = $1
					WHERE link = $2
				`
				dbMock.ExpectExec(regexp.QuoteMeta(query)).
					WithArgs(newBackground, boardLink).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			ExpectedErr: common.ErrBoardNotFound,
		},
		{
			Name: "db error",
			MockSetup: func(dbMock pgxmock.PgxPoolIface) {
				query := `
					UPDATE board_actual
					SET background = $1
					WHERE link = $2
				`
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
