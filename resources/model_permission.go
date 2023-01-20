/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

type Permission struct {
	Key
	Attributes PermissionAttributes `json:"attributes"`
}
type PermissionResponse struct {
	Data     Permission `json:"data"`
	Included Included   `json:"included"`
}

type PermissionListResponse struct {
	Data     []Permission `json:"data"`
	Included Included     `json:"included"`
	Links    *Links       `json:"links"`
}

// MustPermission - returns Permission from include collection.
// if entry with specified key does not exist - returns nil
// if entry with specified key exists but type or ID mismatches - panics
func (c *Included) MustPermission(key Key) *Permission {
	var permission Permission
	if c.tryFindEntry(key, &permission) {
		return &permission
	}
	return nil
}
