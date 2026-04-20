package common

import "errors"

var (
	ErrSectionNotFound             = errors.New("section not found")
	ErrSectionAlreadyExists        = errors.New("section already exists")
	ErrInvalidReferenceSectionData = errors.New("invalid reference section data")
	ErrInvalidSectionData          = errors.New("invalid section data")
	ErrMissingRequiredField        = errors.New("missing required field")
	ErrCannotDeleteBacklog         = errors.New("cannot delete backlog")
	ErrCannotUpdateBacklog         = errors.New("cannot update backlog")
	ErrNotFindAllLinks             = errors.New("not find all links")
	ErrInvalidCardData             = errors.New("invalid card data")
)
