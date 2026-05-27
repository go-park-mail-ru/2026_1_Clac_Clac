.PHONY: go-generate easyjson docs proto

go-generate:
	@go generate ./...

easyjson:
	@easyjson -all -pkg facade/internal/delivery/http/dto
	@easyjson -all -pkg facade/internal/api/dto

docs:
	@swag init -g facade/cmd/api/main.go -o docs --parseDependency

proto:
	@protoc --proto_path=. --go_out=. --go_opt=module=github.com/go-park-mail-ru/2026_1_Clac_Clac \
	--go-grpc_out=. --go-grpc_opt=module=github.com/go-park-mail-ru/2026_1_Clac_Clac \
	proto/board/v1/board.proto proto/section/v1/section.proto proto/card/v1/card.proto \
	proto/appeal/v1/appeal.proto proto/authorization/v1/authorization.proto proto/mail_sender/v1/mail_sender.proto \
	proto/rate_limiter/v1/rate_limiter.proto proto/user/v1/user.proto
