package google

import (
	"os"

	ci "google.golang.org/api/cloudidentity/v1"
)

// GetGroupByID returns information about the group for the given groupID
// from the Instance if already present or calls the google APIs for the same.
func GetGroupByID(groupID string) (*ci.LookupGroupNameResponse, error) {
	// If the information for the group is already present in the instance,
	// return the value from instance.
	if g := instance.GetGroup(groupID); g != nil {
		return g, nil
	}

	svc := instance.CloudIdentityService

	id := groupID + os.Getenv("GOOGLE_GCP_DOMAIN")
	g, err := svc.Groups.Lookup().GroupKeyId(id).Do()
	if err != nil {
		return nil, err
	}

	// Add group information to the instance to avoid repeated calls to GCE for same information.
	instance.AddGroup(groupID, g)
	return g, nil
}

// AddMemberToGroupID adds a member (by emailID) to the group of given groupID
func AddMemberToGroupID(groupID string, emailID string) error {
	// Get Group by groupID
	g, hErr := GetGroupByID(groupID)
	if hErr != nil {
		return hErr
	}

	membership := ci.Membership{
		PreferredMemberKey: &ci.EntityKey{Id: emailID},
		Roles:              []*ci.MembershipRole{{Name: "MEMBER"}},
	}

	svc := instance.CloudIdentityService
	_, err := svc.Groups.Memberships.Create(g.Name, &membership).Do()
	if err != nil {
		return err
	}

	return nil
}

// RemoveMemberFromGroupID removes a member (by emailID) from the group of given groupID
func RemoveMemberFromGroupID(groupID string, emailID string) error {
	name, err := lookupMembershipName(groupID, emailID)
	if err != nil {
		return err
	}

	_, err = instance.CloudIdentityService.Groups.Memberships.Delete(name).Do()
	return err
}
