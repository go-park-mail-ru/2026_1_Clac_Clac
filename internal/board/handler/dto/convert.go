package dto

import (
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/service/dto"
	"github.com/google/uuid"
)

func BoardInfoResponseFromInfo(info serviceDto.BoardInfo) BoardInfoResponse {
	return BoardInfoResponse{
		Link:        info.Link,
		Name:        info.Name,
		Description: info.Description,
		Background:  info.Background,
		CreatedAt:   info.CreatedAt,
	}
}

func ToNewBoardInfo(info CreateBoardRequest) serviceDto.NewBoardInfo {
	return serviceDto.NewBoardInfo{
		Name:        info.Name,
		Description: info.Description,
		Background:  info.Background,
	}
}

func ToUpdateBoardInfo(info UpdateBoardRequest, boardLink uuid.UUID) serviceDto.UpdateBoardInfo {
	return serviceDto.UpdateBoardInfo{
		Link:        boardLink,
		Name:        info.Name,
		Description: info.Description,
		Background:  info.Background,
	}
}
