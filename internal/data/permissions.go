package data

import "time"

type Permissions interface {
	New() Permissions

	Create(user Permission) error
	Upsert(permission Permission) error
	Update(user Permission) error
	UpdateUserId(permission Permission) error
	UpdateParentLevel(permission Permission) error
	Delete(githubId int64, typeTo, link string) error

	JoinsModule() Permissions
	Select() ([]Permission, error)
	Get() (*Permission, error)

	FilterByUserIds(ids ...int64) Permissions
	FilterByGithubIds(githubIds ...int64) Permissions
	FilterByUsernames(usernames ...string) Permissions
	FilterByLinks(links ...string) Permissions
}

type Permission struct {
	RequestId   string    `json:"request_id" db:"request_id" structs:"request_id"`
	UserId      *int64    `json:"user_id" db:"user_id" structs:"user_id"`
	Username    string    `json:"login" db:"username" structs:"username"`
	GithubId    int64     `json:"id" db:"github_id" structs:"github_id"`
	AccessLevel string    `json:"role_name" db:"access_level" structs:"access_level"`
	ParentLevel bool      `json:"-" db:"parent_level" structs:"-"`
	Link        string    `json:"link" db:"link" structs:"link"`
	Type        string    `json:"type" db:"type" structs:"type"`
	CreatedAt   time.Time `json:"created_at" db:"created_at" structs:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at" structs:"-"`
	ExpiresAt   time.Time `json:"expires_at" db:"expires_at" structs:"expires_at"`
	AvatarUrl   string    `json:"avatar_url" db:"-" structs:"-"`
}
