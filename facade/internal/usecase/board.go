package usecase

import (
	grpcclient "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/clients/grpc"
)

type BoardUsecase interface{}

type boardUsecase struct {
	board *grpcclient.BoardClient
}

func NewBoardUsecase(board *grpcclient.BoardClient) BoardUsecase {
	return &boardUsecase{board: board}
}
