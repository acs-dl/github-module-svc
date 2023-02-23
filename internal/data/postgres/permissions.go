package postgres

import (
	"database/sql"
	"fmt"
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

var permissionsColumns = []string{"permissions.request_id", "permissions.user_id", "permissions.username", "permissions.github_id", "permissions.link", "permissions.access_level", "permissions.type"}

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

func (q *PermissionsQ) Update(permission data.Permission) error {
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

func (q *PermissionsQ) UpdateHasChild(permission data.Permission) error {
	query := sq.Update(permissionsTableName).
		Set("has_child", permission.HasChild).
		Where(sq.Eq{
			"github_id": permission.GithubId,
			"link":      permission.Link,
		})

	return q.db.Exec(query)
}

func (q *PermissionsQ) JoinsModule() data.Permissions {
	q.sql = q.sql.
		LeftJoin(fmt.Sprint(usersTableName, " ON ", usersTableName, ".id = ", permissionsTableName, ".user_id"))
	return q
}

func (q *PermissionsQ) Select() ([]data.Permission, error) {
	var result []data.Permission

	err := q.db.Select(&result, q.sql)

	return result, err
}

func (q *PermissionsQ) Upsert(permission data.Permission) error {
	updUsername := fmt.Sprintf("username = '%s'", permission.Username)
	updLevel := fmt.Sprintf("access_level = '%s'", permission.AccessLevel)

	query := sq.Insert(permissionsTableName).SetMap(structs.Map(permission)).
		Suffix(fmt.Sprintf("ON CONFLICT (github_id, link) DO UPDATE SET %s, %s", updUsername, updLevel))

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
	query := sq.Delete(permissionsTableName).Where(
		sq.Eq{"github_id": githubId, "type": typeTo, "link": link})

	result, err := q.db.ExecWithResult(query)
	if err != nil {
		return err
	}

	affectedRows, _ := result.RowsAffected()
	if affectedRows == 0 {
		return errors.New("no user with such github_id")
	}

	return nil
}

func (q *PermissionsQ) FilterByUserIds(ids ...int64) data.Permissions {
	stmt := sq.Eq{permissionsTableName + ".user_id": ids}

	q.sql = q.sql.Where(stmt)

	return q
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
