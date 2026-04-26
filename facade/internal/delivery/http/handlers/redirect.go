package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Redirect sends the client to target with code and message appended as query params.
func Redirect(w http.ResponseWriter, r *http.Request, target string, code int, message string) (http.ResponseWriter, error) {
	u, err := url.Parse(target)
	if err != nil {
		return w, fmt.Errorf("url.Parse: %w", err)
	}

	q := u.Query()
	q.Set("code", strconv.Itoa(code))
	q.Set("message", message)
	u.RawQuery = q.Encode()

	http.Redirect(w, r, u.String(), http.StatusFound)
	return w, nil
}
