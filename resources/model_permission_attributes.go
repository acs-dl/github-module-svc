/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

type PermissionAttributes struct {
	// user id from github
	GithubId int64 `json:"github_id"`
	// link for which was given access
	Link string `json:"link"`
	// level of success for link (pull, push etc..)
	Permission string `json:"permission"`
	// type of link (organization or repository) which was given access
	Type string `json:"type"`
	// user id from identity module
	UserId int64 `json:"user_id"`
	// username from github
	Username string `json:"username"`
}
