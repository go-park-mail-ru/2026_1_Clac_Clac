package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

func Redirect(w http.ResponseWriter, r *http.Request, target string, code int, message string) (http.ResponseWriter, error) {
	const codeParam = "code"
	const messageParam = "message"

	u, err := url.Parse(target)
	if err != nil {
		return w, fmt.Errorf("url.Parse: %w", err)
	}

	q := u.Query()
	q.Set(codeParam, strconv.Itoa(code))
	q.Set(messageParam, message)

	u.RawQuery = q.Encode()

	http.Redirect(w, r, u.String(), http.StatusFound)

	return w, nil
}
