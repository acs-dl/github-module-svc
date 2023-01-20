/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

type UserPermissionAttributes struct {
	// user id from github
	GithubId int64 `json:"github_id"`
	// link for which was given access
	Link string `json:"link"`
	// level of success for link
	Permission string `json:"permission"`
	// type of link (org or repo)
	Type string `json:"type"`
	// username from github
	Username string `json:"username"`
}
