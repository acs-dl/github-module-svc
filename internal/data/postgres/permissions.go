package postgres

import (
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/fatih/structs"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/kit/pgdb"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

const permissionsTableName = "permissions"

type PermissionsQ struct {
	db  *pgdb.DB
	sql sq.SelectBuilder
}

var permissionsColumns = []string{
	permissionsTableName + ".request_id",
	permissionsTableName + ".user_id",
	permissionsTableName + ".username",
	permissionsTableName + ".github_id",
	permissionsTableName + ".link",
	permissionsTableName + ".access_level",
	permissionsTableName + ".type",
	permissionsTableName + ".parent_link",
}

func NewPermissionsQ(db *pgdb.DB) data.Permissions {
	return &PermissionsQ{
		db:  db.Clone(),
		sql: sq.Select(permissionsColumns...).From(permissionsTableName),
	}
}

func (q *PermissionsQ) New() data.Permissions {
	return NewPermissionsQ(q.db)
}

func (q *PermissionsQ) Create(permission data.Permission) error {
	clauses := structs.Map(permission)

	query := sq.Insert(permissionsTableName).SetMap(clauses)

	return q.db.Exec(query)
}

func (q *PermissionsQ) UpdateUsernameAccessLevel(permission data.Permission) error {
	query := sq.Update(permissionsTableName).
		Set("username", permission.Username).
		Set("access_level", permission.AccessLevel).
		Where(
			sq.Eq{"github_id": permission.GithubId, "link": permission.Link})

	return q.db.Exec(query)
}

func (q *PermissionsQ) UpdateUserId(permission data.Permission) error {
	query := sq.Update(permissionsTableName).
		Set("user_id", permission.UserId).
		Where(sq.Eq{"github_id": permission.GithubId})

	return q.db.Exec(query)
}

func (q *PermissionsQ) UpdateHasParent(permission data.Permission) error {
	query := sq.Update(permissionsTableName).
		Set("has_parent", permission.HasParent).
		Where(sq.Eq{
			"github_id": permission.GithubId,
			"link":      permission.Link,
		})

	return q.db.Exec(query)
}

func (q *PermissionsQ) UpdateParentLink(permission data.Permission) error {
	query := sq.Update(permissionsTableName).
		Set("parent_link", permission.ParentLink).
		Where(sq.Eq{
			"github_id": permission.GithubId,
			"link":      permission.Link,
		})

	return q.db.Exec(query)
}

func (q *PermissionsQ) UpdateHasChild(permission data.Permission) error {
	query := sq.Update(permissionsTableName).
		Set("has_child", permission.HasChild).
		Where(sq.Eq{
			"github_id": permission.GithubId,
			"link":      permission.Link,
		})

	return q.db.Exec(query)
}

func (q *PermissionsQ) Select() ([]data.Permission, error) {
	var result []data.Permission

	err := q.db.Select(&result, q.sql)

	return result, err
}

func (q *PermissionsQ) Upsert(permission data.Permission) error {
	updateStmt, args := sq.Update(" ").
		Set("updated_at", time.Now()).
		Set("username", permission.Username).
		Set("access_level", permission.AccessLevel).MustSql()

	query := sq.Insert(permissionsTableName).SetMap(structs.Map(permission)).
		Suffix("ON CONFLICT (github_id, link) DO "+updateStmt, args...)

	return q.db.Exec(query)
}

func (q *PermissionsQ) Get() (*data.Permission, error) {
	var result data.Permission

	err := q.db.Get(&result, q.sql)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &result, err
}

func (q *PermissionsQ) Delete(githubId int64, typeTo, link string) error {
	var deleted []data.Permission

	query := sq.Delete(permissionsTableName).
		Where(sq.Eq{
			"github_id": githubId,
			"type":      typeTo,
			"link":      link,
		}).
		Suffix("RETURNING *")

	err := q.db.Select(&deleted, query)
	if err != nil {
		return err
	}
	if len(deleted) == 0 {
		return errors.Errorf("no rows with `%s` link", link)
	}

	return nil
}

func (q *PermissionsQ) FilterByGithubIds(ids ...int64) data.Permissions {
	stmt := sq.Eq{permissionsTableName + ".github_id": ids}

	q.sql = q.sql.Where(stmt)

	return q
}

func (q *PermissionsQ) FilterByUsernames(usernames ...string) data.Permissions {
	stmt := sq.Eq{permissionsTableName + ".username": usernames}

	q.sql = q.sql.Where(stmt)

	return q
}

func (q *PermissionsQ) FilterByLinks(links ...string) data.Permissions {
	stmt := sq.Eq{permissionsTableName + ".link": links}

	q.sql = q.sql.Where(stmt)

	return q
}

func (q *PermissionsQ) FilterByParentLinks(parentLinks ...string) data.Permissions {
	stmt := sq.Eq{permissionsTableName + ".parent_link": parentLinks}
	if len(parentLinks) == 0 {
		stmt = sq.Eq{permissionsTableName + ".parent_link": nil}
	}

	q.sql = q.sql.Where(stmt)

	return q
}

func (q *PermissionsQ) FilterByHasParent(hasParent bool) data.Permissions {
	q.sql = q.sql.Where(sq.Eq{permissionsTableName + ".has_parent": hasParent})

	return q
}

func (q *PermissionsQ) FilterByGreaterTime(time time.Time) data.Permissions {
	q.sql = q.sql.Where(sq.Gt{permissionsTableName + ".updated_at": time})

	return q
}

func (q *PermissionsQ) FilterByLowerTime(time time.Time) data.Permissions {
	q.sql = q.sql.Where(sq.Lt{permissionsTableName + ".updated_at": time})

	return q
}
