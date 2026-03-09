package api

import "fmt"

const (
	usersGetMethodVkAPI = "https://api.vk.com/method/users.get?fields=email&access_token=%s&v=5.199"
)

type VkAPIUserData struct {
	FirstName string `json:"first_name"`
}

type VkAPIUsersData struct {
	Response []VkAPIUserData `json:"response"`
}

func MakeUserGetVkAPIURL(accessToken string) string {
	return fmt.Sprintf(usersGetMethodVkAPI, accessToken)
}
