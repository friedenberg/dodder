package command_components

import "code.linenisgreat.com/dodder/go/src/alfa/interfaces"

type Checkout struct {
	Delete   bool
	Organize bool
	Edit     bool
}
var _ interfaces.CommandComponentWriter = (*Checkout)(nil)

func (c *Checkout) SetFlagDefinitions(f interfaces.CommandLineFlagDefinitions) {
	f.BoolVar(
		&c.Delete,
		"delete",
		false,
		"delete the zettel and blob after successful checkin",
	)

	f.BoolVar(
		&c.Organize,
		"organize",
		false,
		"open organize",
	)

	f.BoolVar(
		&c.Edit,
		"edit",
		true,
		"create a new empty zettel and open EDITOR or VISUAL for editing and then commit the resulting changes",
	)
}
