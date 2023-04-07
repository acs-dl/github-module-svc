package data

import (
	"time"

	"gitlab.com/distributed_lab/kit/pgdb"
)

type Users interface {
	New() Users

	Upsert(user User) error
	Delete() error
	Select() ([]User, error)
	Get() (*User, error)

	Count() Users
	GetTotalCount() (int64, error)

	FilterById(id *int64) Users
	FilterByUsernames(usernames ...string) Users
	FilterByGithubIds(githubIds ...int64) Users
	FilterByLowerTime(time time.Time) Users
	SearchBy(search string) Users

	Page(pageParams pgdb.OffsetPageParams) Users
}

type User struct {
	Id        *int64    `db:"id" structs:"id,omitempty"`
	Username  string    `json:"login" db:"username" structs:"username"`
	GithubId  int64     `json:"id" db:"github_id" structs:"github_id"`
	AvatarUrl string    `json:"avatar_url" db:"avatar_url" structs:"avatar_url"`
	CreatedAt time.Time `json:"created_at" db:"created_at" structs:"-"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" structs:"-"`
	Submodule *string   `json:"-" db:"-" structs:"-"`
}

type UnverifiedUser struct {
	CreatedAt time.Time `json:"created_at"`
	Module    string    `json:"module"`
	Submodule string    `json:"submodule"`
	ModuleId  string    `json:"module_id"`
	Email     *string   `json:"email,omitempty"`
	Name      *string   `json:"name,omitempty"`
	Phone     *string   `json:"phone,omitempty"`
	Username  *string   `json:"username,omitempty"`
}
