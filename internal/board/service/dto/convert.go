package dto

import (
	repositoryDto "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/repository/dto"
)

func BoardInfoFromEntry(entry repositoryDto.BoardEntry) BoardInfo {
	return BoardInfo{
		Link:        entry.Link,
		Name:        entry.Name,
		Description: entry.Description,
		Background:  entry.Background,
		CreatedAt:   entry.CreatedAt,
	}
}

func ToNewBoardInfo(info NewBoardInfo) repositoryDto.NewBoardInfo {
	return repositoryDto.NewBoardInfo{
		Name:        info.Name,
		Description: info.Description,
		Background:  info.Background,
	}
}

func ToUpdateBoardInfo(info UpdateBoardInfo) repositoryDto.UpdateBoardInfo {
	return repositoryDto.UpdateBoardInfo{
		Link:        info.Link,
		Name:        info.Name,
		Description: info.Description,
		Background:  info.Background,
	}
}
