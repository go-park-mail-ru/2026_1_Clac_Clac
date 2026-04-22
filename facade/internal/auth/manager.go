package auth

import (
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/mail"
)

type Manager struct {
	mailClient pb.MailServiceClient
}

func NewManager(mailClient pb.MailServiceClient) *Manager {
	return &Manager{
		mailClient: mailClient,
	}
}
