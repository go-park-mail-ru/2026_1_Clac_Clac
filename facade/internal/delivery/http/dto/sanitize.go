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

func (r *CreateCardRequest) Sanitize() {
	r.Title = strictSanitizePolicy.Sanitize(r.Title)
}

func (r *UpdateCardRequest) Sanitize() {
	r.Title = strictSanitizePolicy.Sanitize(r.Title)
	r.Description = strictSanitizePolicy.Sanitize(r.Description)
}

func (r *CreateCommentRequest) Sanitize() {
	r.Text = strictSanitizePolicy.Sanitize(r.Text)
}

func (r *UpdateCommentRequest) Sanitize() {
	r.Text = strictSanitizePolicy.Sanitize(r.Text)
}

func (r *CreateSubtaskRequest) Sanitize() {
	r.Description = strictSanitizePolicy.Sanitize(r.Description)
}

func (r *UpdateSubtaskRequest) Sanitize() {
	r.Description = strictSanitizePolicy.Sanitize(r.Description)
}
