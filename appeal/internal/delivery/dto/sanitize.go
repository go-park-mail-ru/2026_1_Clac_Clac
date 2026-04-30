package dto

import "github.com/microcosm-cc/bluemonday"

var (
	strictSanitizePolicy = bluemonday.StrictPolicy()
)

func (r *EntityAppealRequest) Sanitize() {
	r.DisplayName = strictSanitizePolicy.Sanitize(r.DisplayName)
}
