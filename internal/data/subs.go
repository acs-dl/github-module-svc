package data

import "gitlab.com/distributed_lab/kit/pgdb"

type Subs interface {
	New() Subs

	Insert(sub Sub) error
	Upsert(sub Sub) error
	Delete(subId int64, typeTo, link string) error

	Select() ([]Sub, error)
	Get() (*Sub, error)

	FilterByParentIds(parentIds ...int64) Subs
	FilterByLevel(lpath ...string) Subs
	FilterByLowerLevel(parentLpath string) Subs
	FilterByHigherLevel(parentLpath string) Subs
	FilterByLinks(links ...string) Subs
	FilterByIds(ids ...int64) Subs
	SearchBy(search string) Subs

	WithPermissions() Subs
	FilterByGithubIds(githubIds ...int64) Subs
	FilterByUserIds(userIds ...int64) Subs
	FilterByHasParent(level bool) Subs

	ResetFilters() Subs

	DistinctOn(column string) Subs
	OrderBy(columns ...string) Subs

	Count() Subs
	CountWithPermissions() Subs
	GetTotalCount() (int64, error)

	Page(pageParams pgdb.OffsetPageParams) Subs
}

type Sub struct {
	Id          int64  `json:"id" db:"id" structs:"id"`
	Path        string `json:"name" db:"path" structs:"path"`
	Link        string `json:"full_name" db:"subs_link" structs:"link"`
	Type        string `json:"type" db:"subs_type" structs:"type"`
	ParentId    *int64 `json:"parent_id" db:"parent_id" structs:"parent_id"`
	Lpath       string `json:"lpath" db:"lpath" structs:"lpath"`
	*Permission `structs:",omitempty"`
}
