package api

type VkAPIUserData struct {
	FirstName string `json:"first_name"`
}

type VkAPIUsersData struct {
	Response []VkAPIUserData `json:"response"`
}
