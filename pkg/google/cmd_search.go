package google

import (
	"flag"
	"fmt"
	"strings"

	"github.com/aryannr97/unfold/pkg/commands"
	"github.com/aryannr97/unfold/pkg/helpers"
	cloudidentity "google.golang.org/api/cloudidentity/v1"
)

// commandSearchConfig represents the configuration for the search command
type commandSearchConfig struct {
	Members struct {
		ID    *string
		Group *string
	}
	FlagSet *flag.FlagSet
}

// Execute executes the search command
func (c commandSearchConfig) Execute() string {
	if *c.Members.ID != "" {
		found, err := CheckGroupMembershipForEmailIDs(*c.Members.Group, *c.Members.ID)
		if err != nil {
			return fmt.Sprintf("[unfold] %s", err.Error())
		}

		return fmt.Sprintf("[unfold] emailID is found to be %s of the group with membership name %s", helpers.GreenValue(roleNames(found.Roles)), helpers.GreenValue(found.Name))
	}
	return "[unfold] something went wrong"
}

// roleNames joins the membership role names, e.g. "MEMBER,OWNER", or "NONE" when there are none.
func roleNames(roles []*cloudidentity.MembershipRole) string {
	if len(roles) == 0 {
		return "NONE"
	}
	names := make([]string, len(roles))
	for i, r := range roles {
		names[i] = r.Name
	}
	return strings.Join(names, ",")
}

// GetFlagSet returns the flag set for the search command
func (c commandSearchConfig) GetFlagSet() *flag.FlagSet {
	return c.FlagSet
}

// fetchCommandSearchConfig fetches the command search config
func fetchCommandSearchConfig() commandSearchConfig {
	flagSet := flag.NewFlagSet(commands.Search, flag.ContinueOnError)
	return commandSearchConfig{
		Members: struct {
			ID    *string
			Group *string
		}{
			ID:    flagSet.String("id", "", "used to search email in group membership"),
			Group: flagSet.String("g", "", "provide google group id"),
		},
		FlagSet: flagSet,
	}
}
