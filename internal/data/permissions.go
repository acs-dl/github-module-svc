package data

type Permissions interface {
	New() Permissions

	Create(user Permission) error
	Upsert(permission Permission) error
	Update(user Permission) error
	UpdateUserId(permission Permission) error
	Delete(githubId int64, typeTo, link string) error

	JoinsModule() Permissions
	Select() ([]Permission, error)
	Get() (*Permission, error)
	FilterByUserIds(ids ...int64) Permissions
}

type Permission struct {
	RequestId  string `json:"request_id" db:"request_id" structs:"request_id"`
	UserId     *int64 `json:"user_id" db:"user_id" structs:"user_id"`
	Username   string `json:"login" db:"username" structs:"username"`
	GithubId   int64  `json:"id" db:"github_id" structs:"github_id"`
	Permission string `json:"role_name" db:"permission" structs:"permission"`
	Link       string `json:"link" db:"link" structs:"link"`
	Type       string `json:"type_to" db:"type_to" structs:"type_to"`
}
