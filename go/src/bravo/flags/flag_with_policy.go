package flags

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/flag_policy"
)

func MakeWithPolicy(
	fp flag_policy.FlagPolicy,
	stringer func() string,
	set func(string) error,
	reset func(),
) FlagWithPolicy {
	return FlagWithPolicy{
		FlagPolicy: fp,
		stringer:   stringer,
		set:        set,
		reset:      reset,
	}
}

type FlagWithPolicy struct {
	flag_policy.FlagPolicy
	stringer func() string
	set      func(string) error
	reset    func()
}

func (flag FlagWithPolicy) Set(v string) (err error) {
	if flag.FlagPolicy == flag_policy.FlagPolicyReset {
		flag.reset()
	}

	if err = flag.set(v); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (flag FlagWithPolicy) String() string {
	if flag.stringer == nil {
		return "nil"
	} else {
		return flag.stringer()
	}
}
