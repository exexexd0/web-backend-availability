package role

// Role — роль пользователя в JWT.
// Гость роли не имеет: у него нет токена.
type Role string

const (
	Creator   Role = "creator"
	Moderator Role = "moderator"
)

func FromModerator(isModerator bool) Role {
	if isModerator {
		return Moderator
	}
	return Creator
}
