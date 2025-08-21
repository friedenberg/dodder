package command

import (
	"flag"
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
	"code.linenisgreat.com/dodder/go/src/foxtrot/repo_config_cli"
)

type Request struct {
	interfaces.Context
	repo_config_cli.Config
	*flag.FlagSet
	*Args
}

type consumedArg struct {
	name, value string
}

func (arg consumedArg) String() string {
	if arg.name == "" {
		return fmt.Sprintf("%q", arg.value)
	} else {
		return fmt.Sprintf("%s:%q", arg.name, arg.value)
	}
}

type Args struct {
	interfaces.Context
	args []string
	argi int

	consumed []consumedArg
}

func MakeRequest(
	ctx interfaces.Context,
	config repo_config_cli.Config,
	flagSet *flag.FlagSet,
) Request {
	return Request{
		Context: ctx,
		Config:  config,
		FlagSet: flagSet,
		Args: &Args{
			Context: ctx,
			args:    flagSet.Args(),
		},
	}
}

func (req *Args) PeekArgs() []string {
	args := req.args[req.argi:]
	return args
}

func (req *Args) PopArgs() []string {
	args := req.PeekArgs()

	for _, arg := range args {
		req.consumed = append(req.consumed, consumedArg{value: arg})
	}

	req.argi += len(args)
	return args
}

func (req *Args) PopArgsAsMutableSet() collections_value.MutableSet[string] {
	args := req.PeekArgs()
	set := collections_value.MakeMutableSet(
		quiter.StringKeyer,
	)

	for _, arg := range args {
		set.Add(arg)
		req.consumed = append(req.consumed, consumedArg{value: arg})
	}

	req.argi += len(args)
	return set
}

func (req *Args) RemainingArgCount() int {
	return len(req.args[req.argi:])
}

// TODO use pools?
func PopRequestArg[
	T interfaces.Stringer,
	Ptr interfaces.StringerSetterPtr[T],
](req *Args, name string) Ptr {
	arg := req.PopArg(name)
	var argValue T

	if err := (Ptr(&argValue)).Set(arg); err != nil {
		errors.ContextCancelWithBadRequestError(req, err)
	}

	return &argValue
}

func (req *Args) PopArg(name string) string {
	if req.RemainingArgCount() == 0 {
		errors.ContextCancelWithBadRequestf(
			req,
			"expected positional argument (%d) %s, but only received %q",
			req.argi+1,
			name,
			req.consumed,
		)
	}

	value := req.args[req.argi]
	req.consumed = append(req.consumed, consumedArg{name: name, value: value})
	req.argi++
	return value
}

func (req *Args) PopArgOrDefault(name, defaultArg string) string {
	if req.RemainingArgCount() == 0 {
		return defaultArg
	}

	value := req.args[req.argi]
	req.consumed = append(req.consumed, consumedArg{name: name, value: value})
	req.argi++

	return value
}

func (req *Args) AssertNoMoreArgs() {
	if req.RemainingArgCount() > 0 {
		errors.ContextCancelWithBadRequestf(
			req,
			"expected no more arguments, but have %q",
			req.PopArgs(),
		)
	}
}

func (req *Args) LastArg() (arg string, ok bool) {
	if req.RemainingArgCount() > 0 {
		ok = true
		arg = req.PopArgs()[req.RemainingArgCount()-1]
	}

	return
}
