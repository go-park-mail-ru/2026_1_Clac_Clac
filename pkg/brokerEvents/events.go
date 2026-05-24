package brokerEvents

type BoardUpdateEvent struct {
	BoardLink string `json:"board_link"`
	UserLink  string `json:"user_link"`
	Action    string `json:"action"`
	Data      any    `json:"data"`
}
