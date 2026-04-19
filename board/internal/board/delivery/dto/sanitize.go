package dto

import "github.com/microcosm-cc/bluemonday"

var (
	strictSanitizePolicy = bluemonday.StrictPolicy()
	ugcSanitizePolicy    = bluemonday.UGCPolicy()
)

func (b *CreateBoardRequest) Sanitize() {
	b.Name = strictSanitizePolicy.Sanitize(b.Name)
	b.Description = ugcSanitizePolicy.Sanitize(b.Description)
	b.Background = strictSanitizePolicy.Sanitize(b.Background)
}

func (b *UpdateBoardRequest) Sanitize() {
	b.Name = strictSanitizePolicy.Sanitize(b.Name)
	b.Description = ugcSanitizePolicy.Sanitize(b.Description)
	b.Background = strictSanitizePolicy.Sanitize(b.Background)
}
