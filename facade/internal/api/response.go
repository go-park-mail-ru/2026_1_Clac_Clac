package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/api/dto"
	"github.com/mailru/easyjson"
	"github.com/mailru/easyjson/jwriter"
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

type Response = dto.Response
type ErrorResponse = dto.ErrorResponse

// Ответ для 200 статуса, всегда должен содержать данные
//
//	@Description	Успешный ответ с данными
type OkResponse[T any] struct {
	Status string `json:"status"`
	Data   T      `json:"data"`
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
func respond(w http.ResponseWriter, statusCode int, response any) (http.ResponseWriter, error) {
	if marshaler, ok := response.(easyjson.Marshaler); ok {
		var jw jwriter.Writer
		marshaler.MarshalEasyJSON(&jw)

		if jw.Error != nil {
			return w, fmt.Errorf("easyjson marshal error: %w", jw.Error)
		}

		SetContentType(w, MIMEApplicationJSON)
		w.WriteHeader(statusCode)

		_, err := jw.DumpTo(w)
		if err != nil {
			return w, fmt.Errorf("cannot write easyjson: %w", err)
		}

		return w, nil
	}

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
	response := &dto.Response{
		Status: status,
	}

	return respond(w, statusCode, response)
}

// Всегда возвращает 200-ку, принимает любые данные, которые можно маршалить
func RespondOk[T any](w http.ResponseWriter, data T) (http.ResponseWriter, error) {
	if m, ok := any(data).(easyjson.Marshaler); ok {
		var innerJW jwriter.Writer
		m.MarshalEasyJSON(&innerJW)

		if innerJW.Error != nil {
			return w, fmt.Errorf("inner marshal error: %w", innerJW.Error)
		}

		rawBytes, _ := innerJW.BuildBytes()

		fastResp := &dto.OkResponseFast{
			Status: StatusOK,
			Data:   rawBytes,
		}

		return respond(w, http.StatusOK, fastResp)
	}

	response := OkResponse[T]{
		Status: StatusOK,
		Data:   data,
	}

	return respond(w, http.StatusOK, response)
}

func RespondCreated[T any](w http.ResponseWriter, data T) (http.ResponseWriter, error) {
	if m, ok := any(data).(easyjson.Marshaler); ok {
		var innerJw jwriter.Writer
		m.MarshalEasyJSON(&innerJw)

		if innerJw.Error != nil {
			return w, fmt.Errorf("inner marshal error: %w", innerJw.Error)
		}

		rawBytes, _ := innerJw.BuildBytes()
		fastResp := &dto.OkResponseFast{
			Status: StatusOK,
			Data:   rawBytes,
		}
		return respond(w, http.StatusCreated, fastResp)
	}

	response := OkResponse[T]{
		Status: StatusOK,
		Data:   data,
	}
	return respond(w, http.StatusCreated, response)
}

// Ответ ошибкой, надо указать код ошибки и сообщение.
// Код ошибки также установится в HTTP-ответ, поэтому надо использовать валидные коды
func RespondError(w http.ResponseWriter, errorCode int, message string) (http.ResponseWriter, error) {
	response := &dto.ErrorResponse{
		Response: dto.Response{
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
