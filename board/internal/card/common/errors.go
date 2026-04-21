package common

import "errors"

var (
	ErrCardNotFound               = errors.New("card not found")
	ErrInvalidReferenceCardData   = errors.New("invalid reference card data")
	ErrInvalidCardData            = errors.New("invalid card data")
	ErrMissingRequiredField       = errors.New("missing required field")
	ErrCannotSkipMandatorySection = errors.New("cannot skip mandatory section")
	ErrSectionNotFound            = errors.New("section not found")
	ErrCardAlreadyExists          = errors.New("card already exists")
)
