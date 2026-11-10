package azure

import (
	"flag"
	"fmt"

	"github.com/aryannr97/unfold/pkg/commands"
	"github.com/aryannr97/unfold/pkg/helpers"
)

// commandSearchConfig represents the configuration for the search command
type commandSearchConfig struct {
	AudienceOpts struct {
		ID    *string
		Offer *string
	}
	FlagSet *flag.FlagSet
}

// Execute executes the search command
func (c commandSearchConfig) Execute() string {
	id, name := *c.AudienceOpts.ID, *c.AudienceOpts.Offer
	switch {
	case id == "":
		return "[unfold] id cannot be empty"
	case !helpers.IsGUID(id):
		return "[unfold] id must be a GUID"
	case name == "":
		return "[unfold] offer cannot be empty"
	}
	offer, ok := config.Offers[name]
	if !ok {
		return fmt.Sprintf("[unfold] unknown offer %q", name)
	}
	return fmt.Sprintf("[unfold] %s", Search(id, offer.ProductDurableID))
}

// GetFlagSet returns the flag set for the search command
func (c commandSearchConfig) GetFlagSet() *flag.FlagSet {
	return c.FlagSet
}

// fetchCommandSearchConfig fetches the command search config
func fetchCommandSearchConfig() commandSearchConfig {
	flagSet := flag.NewFlagSet(commands.Search, flag.ContinueOnError)
	return commandSearchConfig{
		AudienceOpts: struct {
			ID    *string
			Offer *string
		}{
			ID:    flagSet.String("id", "", "used to search resource in private audience of offer"),
			Offer: flagSet.String("o", "", "provide a valid azure offer name"),
		},
		FlagSet: flagSet,
	}
}
