package dto

import "github.com/mailru/easyjson"

// Ответ для ошибки, всегда содержит код ошибки и сообщение
// @Description	Структура сообщения об ошибке
type ErrorResponse struct {
	Response
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Обертка для успешных (200 OK) ответов API
// Использует easyjson.RawMessage для прямой вставки предварительно сгенерированного JSON
// @Description Успешный ответ сервера с данными
type OkResponseFast struct {
	Status string              `json:"status" example:"ok"`
	Data   easyjson.RawMessage `json:"data" swaggertype:"object"`
}

// Определяет структуру всех ответов API
// Все ответы имеют единое поле status
//
//	@Description	Базовый ответ
type Response struct {
	Status string `json:"status"`
}
