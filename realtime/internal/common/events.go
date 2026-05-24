package common

type BoardUpdateEvent struct {
	BoardLink  string `json:"board_link"`
	EntityType string `json:"entity_type"`
	EntityLink string `json:"entity_link"`
	Action     string `json:"action"`
}
