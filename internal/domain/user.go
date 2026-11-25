package domain

import (
	"strings"
	"time"
)

type AccountStatus string

const (
	AccountActive  AccountStatus = "ACTIVE"
	AccountBlocked AccountStatus = "BLOCKED"
)

// Role — роль користувача. Administrator із діаграми класів реалізовано
// як роль (композиція замість наслідування, що ідіоматично для Go).
type Role string

const (
	RoleGuest         Role = "GUEST"
	RoleUser          Role = "USER"
	RoleAdministrator Role = "ADMINISTRATOR"
)

type Theme string

const (
	ThemeLight Theme = "LIGHT"
	ThemeDark  Theme = "DARK"
)

type Platform string

const (
	PlatformIOS     Platform = "IOS"
	PlatformAndroid Platform = "ANDROID"
	PlatformWeb     Platform = "WEB"
)

func (p Platform) IsValid() bool {
	return p == PlatformIOS || p == PlatformAndroid || p == PlatformWeb
}

// User — обліковий запис (Рис. 3.1.1).
type User struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Surname       string        `json:"surname"`
	Email         string        `json:"email"`
	PasswordHash  string        `json:"-"`
	GoogleID      string        `json:"-"`
	Status        AccountStatus `json:"status"`
	Role          Role          `json:"role"`
	InterfaceLang string        `json:"interfaceLang"`
	Theme         Theme         `json:"theme"`
	CreatedAt     time.Time     `json:"createdAt"`
}

func (u *User) IsBlocked() bool       { return u.Status == AccountBlocked }
func (u *User) IsAdministrator() bool { return u.Role == RoleAdministrator }

func (u *User) UpdateProfile(name, surname string) error {
	if strings.TrimSpace(name) == "" {
		return Invalid("name is required")
	}
	u.Name, u.Surname = strings.TrimSpace(name), strings.TrimSpace(surname)
	return nil
}

// Session — сеанс користувача. Зберігається лише хеш токена.
type Session struct {
	ID        string
	UserID    string
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
}

func (s *Session) IsExpired(now time.Time) bool { return !now.Before(s.ExpiresAt) }

// Device — зареєстрований пристрій для push-сповіщень (PushDevice).
type Device struct {
	ID           string    `json:"id"`
	UserID       string    `json:"userId"`
	FCMToken     string    `json:"fcmToken"`
	Platform     Platform  `json:"platform"`
	RegisteredAt time.Time `json:"registeredAt"`
}
