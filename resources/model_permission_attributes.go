/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

type PermissionAttributes struct {
	AccessLevel Role `json:"access_level"`
	// link for which was given access
	Link string `json:"link"`
	// user id from gitlab
	ModuleId int64 `json:"module_id"`
	// user name from gitlab
	Name string `json:"name"`
	// type of link for which was given access (group or project)
	Type string `json:"type"`
	// user id from identity module
	UserId int64 `json:"user_id"`
	// username from gitlab
	Username string `json:"username"`
}
