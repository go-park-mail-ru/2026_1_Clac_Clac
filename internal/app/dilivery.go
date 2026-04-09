package app

import (
	auth "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/handler"
	board "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/handler"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	profile "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/handler"
	section "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/section/handler"
)

type Dilivery struct {
	Auth    *auth.Handler
	Profile *profile.Handler
	Board   *board.Handler
	Section *section.Handler
}

func NewDilivery(m *Manager, conf *config.Config) *Dilivery {
	authDeps := auth.Deps{
		Srv: m.Auth,

		MaxLenPassword:  conf.Auth.Handler.MaxLenPassword,
		MinLenPassword:  conf.Auth.Handler.MinLenPassword,
		SessionLifetime: conf.Auth.Handler.SessionLifetime,
	}

	profileDeps := profile.Deps{
		Srv:             m.Profile,
		ValidExtensions: conf.S3Avatars.ValidExtensions,

		SiganatureTypeBytes:   conf.Profile.Handler.SiganatureTypeBytes,
		MaxLenNameUser:        conf.Profile.Handler.MaxLenNameUser,
		MaxLenDescriptionUser: conf.Profile.Handler.MaxLenDescriptionUser,
		MaxReadBytes:          conf.Profile.Handler.MaxReadBytes,
	}

	sectionDeps := section.Deps{
		Srv: m.Section,

		MaxQuantityTasks:  conf.Section.Handler.MaxQuantityTasks,
		MinQuantityTasks:  conf.Section.Handler.MinQuantityTasks,
		MaxLenNameSection: conf.Section.Handler.MaxLenNameSection,
	}

	return &Dilivery{
		Auth:    auth.NewHandler(authDeps),
		Profile: profile.NewHandler(profileDeps),
		Board:   board.NewHandler(m.Board),
		Section: section.NewHandler(sectionDeps),
	}
}
