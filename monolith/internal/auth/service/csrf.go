package service

const (
	csrfTokenExpireInHours                  = 24
	csrfTokenExpireTimeConvertationBase     = 10
	csrfTokenExpireTimeConvertationTypeSize = 64 // int64
	csrfTokenPartsCount                     = 2  // hash:expire time
)
