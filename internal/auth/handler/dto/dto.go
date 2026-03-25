package dto

type RegistraionInfoRequest struct {
	Name     string
	Email    string
	Password string
}

type LoginInfoRequest struct {
	Email    string
	Password string
}
