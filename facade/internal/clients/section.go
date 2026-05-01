package clients

import (
	pb "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/proto/section/v1"
	"google.golang.org/grpc"
)

type Section struct {
	client pb.SectionServiceClient
}

func NewSectionClient(connection *grpc.ClientConn) *Section {
	return &Section{
		client: pb.NewSectionServiceClient(connection),
	}
}
