package flag

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/flag_policy"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
)

type (
	FlagSet = flags.FlagSet
	Value   = interfaces.FlagValue
)

func Make(
	fp flag_policy.FlagPolicy,
	stringer func() string,
	set func(string) error,
	reset func(),
) Flag {
	return Flag{
		FlagPolicy: fp,
		stringer:   stringer,
		set:        set,
		reset:      reset,
	}
}

type Flag struct {
	flag_policy.FlagPolicy
	stringer func() string
	set      func(string) error
	reset    func()
}

func (f Flag) Set(v string) (err error) {
	if f.FlagPolicy == flag_policy.FlagPolicyReset {
		f.reset()
	}

	if err = f.set(v); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (f Flag) String() string {
	if f.stringer == nil {
		return "nil"
	} else {
		return f.stringer()
	}
}
