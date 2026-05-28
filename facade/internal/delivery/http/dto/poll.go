package dto

// CreatePollRequest содержит данные для создания покер-комнаты.
//
//	@Description	Данные для создания комнаты Planning Poker
type CreatePollRequest struct {
	CardLinks []string `json:"card_links"`
	Invitees  []string `json:"invitees"`
}

// VotePollRequest содержит оценку участника.
//
//	@Description	Оценка (story points) участника голосования
type VotePollRequest struct {
	Points int `json:"points"`
}
