package dto

import "github.com/microcosm-cc/bluemonday"

var (
	strictSanitizePolicy = bluemonday.StrictPolicy()
)

func (r *RegisterRequest) Sanitize() {
	r.DisplayName = strictSanitizePolicy.Sanitize(r.DisplayName)
}

func (u *UpdateProfileRequest) Sanitize() {
	u.DisplayName = strictSanitizePolicy.Sanitize(u.DisplayName)
}
