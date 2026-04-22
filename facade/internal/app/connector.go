package app

import (
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/config"
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/mail"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Connector struct {
	MailSenderClient pb.MailServiceClient
}

func NewConnector(conf *config.Config, logger *zerolog.Logger) *Connector {
	connection, err := grpc.NewClient(conf.MailSenderURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal().Msgf("Can not connect with MailService: %s", err.Error())
	}

	mailClient := pb.NewMailServiceClient(connection)

	return &Connector{
		MailSenderClient: mailClient,
	}
}
