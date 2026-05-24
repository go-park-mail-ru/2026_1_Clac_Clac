package brokerEvents

type BoardUpdateEvent struct {
	BoardLink string `json:"board_link"`
	UserLink  string `json:"user_link"`
	Data      any    `json:"data"`
}
