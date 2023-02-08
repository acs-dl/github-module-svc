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

	FilterByIds(ids ...*int64) Users
	SearchBy(search string) Users

	Page(pageParams pgdb.OffsetPageParams) Users
}

type User struct {
	Id        *int64    `db:"id" structs:"id,omitempty"`
	Username  string    `json:"login" db:"username" structs:"username"`
	GithubId  int64     `json:"id" db:"github_id" structs:"github_id"`
	AvatarUrl string    `json:"avatar_url" db:"avatar_url" structs:"avatar_url"`
	CreatedAt time.Time `json:"created_at" db:"created_at" structs:"created_at"`
}
