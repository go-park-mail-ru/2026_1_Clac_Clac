package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// HTTP заголовки
const (
	HeaderContentType = "Content-Type"
)

// MIME типы
const (
	MIMEApplicationJSON = "application/json"
	MIMETextPlain       = "text/plain"
)

// Статусы ответов
const (
	StatusOK    = "ok"
	StatusError = "error"
)

// Response определяет базовую структуру всех ответов API
//
//	@Description	Базовый ответ
type Response struct {
	Status string `json:"status" example:"ok"`
}

// OkResponse ответ для успешных операций (200 OK), всегда содержит данные
//
//	@Description	Успешный ответ с данными
type OkResponse[T any] struct {
	Response
	Data T `json:"data"`
}

// ErrorResponse ответ при возникновении ошибки, содержит код и сообщение
//
//	@Description	Структура сообщения об ошибке
type ErrorResponse struct {
	Response
	Code    int    `json:"code" example:"404"`
	Message string `json:"message" example:"not found"`
}

// Устанавливает заголовок ответа Content-Type в переданное значение.
// Функция не проверяет переданное значение, поэтому лучше использовать константы MIME типов.
// Если заголовок Content-Type уже был установлен ранее, то функция ничего не изменит
func SetContentType(w http.ResponseWriter, contentType string) {
	if header := w.Header().Get(HeaderContentType); header == "" {
		w.Header().Set(HeaderContentType, contentType)
	}
}

// Принимает response T, переводит его в json и отправляет.
// Возвращает ошибку, если маршалинг или запись в ответ не удалась.
// Устанавливает Content-Type и записывает статус
func respond[T any](w http.ResponseWriter, statusCode int, response T) (http.ResponseWriter, error) {
	bytes, err := json.Marshal(response)
	if err != nil {
		return w, fmt.Errorf("cannot marshal object to json: %w", err)
	}

	SetContentType(w, MIMEApplicationJSON)
	w.WriteHeader(statusCode)

	_, err = w.Write(bytes)
	if err != nil {
		return w, fmt.Errorf("cannot write json: %w", err)
	}

	return w, nil
}

// Отправляет JSON с единственным полем status
func Respond(w http.ResponseWriter, statusCode int, status string) (http.ResponseWriter, error) {
	response := Response{
		Status: status,
	}

	return respond(w, statusCode, response)
}

// Всегда возвращает 200-ку, принимает любые данные, которые можно маршалить
func RespondOk[T any](w http.ResponseWriter, data T) (http.ResponseWriter, error) {
	response := OkResponse[T]{
		Response: Response{
			Status: StatusOK,
		},
		Data: data,
	}

	return respond(w, http.StatusOK, response)
}

// Ответ ошибкой, надо указать код ошибки и сообщение.
// Код ошибки также установится в HTTP-ответ, поэтому надо использовать валидные коды
func RespondError(w http.ResponseWriter, errorCode int, message string) (http.ResponseWriter, error) {
	response := ErrorResponse{
		Response: Response{
			Status: StatusError,
		},
		Code:    errorCode,
		Message: message,
	}

	return respond(w, errorCode, response)
}

// Если err != nil возвращает клиенту 500-ку.
// Небольшой хелпер, чтобы код стал чище
func HandleError(w http.ResponseWriter, err error) error {
	const errorMessage = "Internal Server Error"

	if err != nil {
		http.Error(w, errorMessage, http.StatusInternalServerError)
	}

	return err
}
