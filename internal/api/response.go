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

// Устанавливает заголовок ответа Content-Type в переданное значение.
// Вызывать функцию надо ДО записи заголовка в ответ.
// Функция не проверяет значение, поэтому лучше использовать константы MIME типов.
// Если заголовок Content-Type уже был установлен ранее, то функция ничего не изменит
func SetContentType(w http.ResponseWriter, contentType string) {
	if header := w.Header().Get(HeaderContentType); header == "" {
		w.Header().Set(HeaderContentType, contentType)
	}
}

// Записывает переданную строку в ответ, устанавливает статус ответа.
// В качетсве Content-Type устанавливает text/plain.
// Данная функция также записывает заголовок.
// Вернет ошибку, если запись не удалась
func RespondString(w http.ResponseWriter, statusCode int, content string) error {
	SetContentType(w, MIMETextPlain)
	w.WriteHeader(statusCode)

	_, err := w.Write([]byte(content))
	if err != nil {
		return fmt.Errorf("cannot write string: %w", err)
	}

	return nil
}

// Записывает переданное значение в ответ, устанавливает статус ответа.
// В качетсве Content-Type устанавливает application/json.
// Данная функция также записывает заголовок.
// Вернет ошибку, если запись не удалась
func RespondJSON(w http.ResponseWriter, statusCode int, payload any) error {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("cannot marshal object to json: %w", err)
	}

	SetContentType(w, MIMEApplicationJSON)
	w.WriteHeader(statusCode)

	_, err = w.Write(bytes)
	if err != nil {
		return fmt.Errorf("cannot write json: %w", err)
	}

	return nil
}
