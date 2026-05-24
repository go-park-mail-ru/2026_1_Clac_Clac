package dto

import "github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/common"

type BoardUpdateInfo struct {
	Type    string                  `json:"type"`
	Payload common.BoardUpdateEvent `json:"payload"`
}
