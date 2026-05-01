package clients

import (
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/card/v1"
	"google.golang.org/grpc"
)

type Card struct {
	client pb.CardServiceClient
}

func NewCardClient(connection *grpc.ClientConn) *Card {
	return &Card{
		client: pb.NewCardServiceClient(connection),
	}
}
