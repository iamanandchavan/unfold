package google

import (
	"errors"
	"net/http"

	"google.golang.org/api/cloudidentity/v1"
	"google.golang.org/api/googleapi"
)

// errMemberNotFound is returned when the email is not a member of the group.
var errMemberNotFound = errors.New("member not found")

// lookupMembershipName returns the resource name of the membership of emailID in
// the group with given groupID, with a single lookup call instead of listing members.
func lookupMembershipName(groupID string, emailID string) (string, error) {
	g, err := GetGroupByID(groupID)
	if err != nil {
		return "", err
	}

	res, err := instance.CloudIdentityService.Groups.Memberships.Lookup(g.Name).MemberKeyId(emailID).Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == http.StatusNotFound {
			return "", errMemberNotFound
		}
		return "", err
	}
	return res.Name, nil
}

// CheckGroupMembershipForEmailIDs returns the membership of the given emailID in
// the google-group with given groupID, or an error if it is not a member.
func CheckGroupMembershipForEmailIDs(groupID string, emailID string) (*cloudidentity.Membership, error) {
	name, err := lookupMembershipName(groupID, emailID)
	if err != nil {
		return nil, err
	}
	return instance.CloudIdentityService.Groups.Memberships.Get(name).Do()
}
