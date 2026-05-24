package dto

import "github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/common"

type BoardUpdateInfo struct {
	Type    string
	Payload common.BoardUpdateEvent
}
