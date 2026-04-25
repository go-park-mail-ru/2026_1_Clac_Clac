package app

import (
	appeal "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/handler"
	auth "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/handler"
	board "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/handler"
	card "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/card/handler"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	profile "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/handler"
	section "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/section/handler"
)

type Dilivery struct {
	Auth    *auth.Handler
	Profile *profile.Handler
	Board   *board.BoardHandler
	Section *section.Handler
	Card    *card.Handler
	Appeal  *appeal.Handler
}

func NewDilivery(m *Manager, conf *config.Config) *Dilivery {
	authConfig := auth.Config{
		MaxLenPassword:  conf.Auth.Handler.MaxLenPassword,
		MinLenPassword:  conf.Auth.Handler.MinLenPassword,
		SessionLifetime: conf.Auth.Handler.SessionLifetime,
	}

	profileConfig := profile.Config{
		ValidExtensions: conf.S3Avatars.ValidExtensions,

		SiganatureTypeBytes:   conf.Profile.Handler.SiganatureTypeBytes,
		MaxLenNameUser:        conf.Profile.Handler.MaxLenNameUser,
		MaxLenDescriptionUser: conf.Profile.Handler.MaxLenDescriptionUser,
		MaxReadBytes:          conf.Profile.Handler.MaxReadBytes,
	}

	sectionConfig := section.Config{
		MaxQuantityTasks:  conf.Section.Handler.MaxQuantityTasks,
		MinQuantityTasks:  conf.Section.Handler.MinQuantityTasks,
		MaxLenNameSection: conf.Section.Handler.MaxLenNameSection,
	}

	cardConfig := card.Config{
		MaxLenTitle:       conf.Card.Handler.MaxLenTitle,
		MaxLenDescription: conf.Card.Handler.MaxLenDescription,
	}

	boardConfig := board.Config{
		MultipartBackgroundFileKey: conf.Board.Handler.MultipartBackgroundFileKey,
		MaxBackgroundSize:          conf.Board.Handler.MaxBackgroundSize,
	}

	return &Dilivery{
		Auth:    auth.NewHandler(m.Auth, authConfig),
		Profile: profile.NewHandler(m.Profile, profileConfig),
		Board:   board.NewHandler(m.Board, boardConfig),
		Section: section.NewHandler(m.Section, sectionConfig),
		Card:    card.NewHandler(m.Card, cardConfig),
		Appeal:  appeal.NewHandler(m.Appeal),
	}
}
