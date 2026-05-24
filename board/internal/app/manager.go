package app

import (
	board "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/service"
	card "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/card/service"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/poll"
	section "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/section/service"
	rbac "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/boardRbac"
	s3 "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/s3"
)

type Manager struct {
	Board             *board.Service
	Section           *section.Service
	Card              *card.Service
	PollStore         *poll.PollStore
	PermissionChecker rbac.Service
}

func NewManager(store *Store, conf *config.Config) *Manager {
	permissionChecker := rbac.NewCachedService(store.PermissionChecker, store.RedisClient)

	pollStore := poll.NewPollStore()

	configCard := card.Config{
		BaseURLAttachment: s3.GetURL(conf.S3.Endpoint, conf.S3.CardsAttachmentBucket),
	}

	return &Manager{
		Board:             board.NewService(store.Board, permissionChecker),
		Section:           section.NewService(store.Section, permissionChecker),
		Card:              card.NewService(store.Card, permissionChecker, pollStore, store.Publisher, configCard),
		PollStore:         pollStore,
		PermissionChecker: permissionChecker,
	}
}
