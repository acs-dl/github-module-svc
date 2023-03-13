package data

import (
	"gitlab.com/distributed_lab/kit/pgdb"
	"time"
)

type Users interface {
	New() Users

	Upsert(user User) error
	Delete(githubId int64) error

	Select() ([]User, error)
	Get() (*User, error)

	Count() Users
	GetTotalCount() (int64, error)

	GetById(id int64) (*User, error)
	GetByUsername(username string) (*User, error)
	GetByGithubId(githubId int64) (*User, error)

	FilterByTime(time time.Time) Users
	FilterById(id *int64) Users
	SearchBy(search string) Users

	Page(pageParams pgdb.OffsetPageParams) Users

	ResetFilters() Users
}

type User struct {
	Id        *int64    `db:"id" structs:"id,omitempty"`
	Username  string    `json:"login" db:"username" structs:"username"`
	GithubId  int64     `json:"id" db:"github_id" structs:"github_id"`
	AvatarUrl string    `json:"avatar_url" db:"avatar_url" structs:"avatar_url"`
	CreatedAt time.Time `json:"created_at" db:"created_at" structs:"-"`
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
