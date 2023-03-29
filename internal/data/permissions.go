package data

import "time"

type Permissions interface {
	New() Permissions

	Create(user Permission) error
	Upsert(permission Permission) error
	UpdateUsernameAccessLevel(user Permission) error
	UpdateUserId(permission Permission) error
	UpdateHasParent(permission Permission) error
	UpdateHasChild(permission Permission) error
	UpdateParentLink(permission Permission) error
	Delete(githubId int64, typeTo, link string) error

	Select() ([]Permission, error)
	Get() (*Permission, error)

	FilterByGithubIds(githubIds ...int64) Permissions
	FilterByUsernames(usernames ...string) Permissions
	FilterByLinks(links ...string) Permissions
	FilterByGreaterTime(time time.Time) Permissions
	FilterByLowerTime(time time.Time) Permissions
	FilterByParentLinks(parentLinks ...string) Permissions
	FilterByHasParent(hasParent bool) Permissions
}

type Permission struct {
	RequestId   string    `json:"request_id" db:"request_id" structs:"request_id"`
	UserId      *int64    `json:"user_id" db:"user_id" structs:"user_id"`
	Username    string    `json:"login" db:"username" structs:"username"`
	GithubId    int64     `json:"id" db:"github_id" structs:"github_id"`
	AccessLevel string    `json:"role_name" db:"access_level" structs:"access_level"`
	HasParent   bool      `json:"-" db:"has_parent" structs:"-"`
	HasChild    bool      `json:"-" db:"has_child" structs:"-"`
	Link        string    `json:"link" db:"link" structs:"link"`
	Type        string    `json:"type" db:"type" structs:"type"`
	ParentLink  *string   `json:"-" db:"parent_link" structs:"parent_link"`
	CreatedAt   time.Time `json:"created_at" db:"created_at" structs:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at" structs:"-"`
	ExpiresAt   time.Time `json:"expires_at" db:"expires_at" structs:"expires_at"`
	AvatarUrl   string    `json:"avatar_url" db:"-" structs:"-"`
}
