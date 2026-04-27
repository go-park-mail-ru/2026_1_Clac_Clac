package dto

import (
<<<<<<<< HEAD:board/internal/board/delivery/dto/convert.go
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/board/service/dto"
========
	serviceDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/board/service/dto"
>>>>>>>> feat/add-facade:monolith/internal/board/handler/dto/convert.go
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
