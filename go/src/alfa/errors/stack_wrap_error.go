package errors

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/stack_frame"
)

type stackWrapError struct {
	ExtraData string
	stack_frame.Frame
	error

	next *stackWrapError
}

func (err *stackWrapError) checkCycle() {
	slow := err
	fast := err

	for fast != nil {
		if slow == fast && slow != err {
			panic("cycle detected!")
		}

		slow = slow.next
		fast = fast.next

		if fast != nil {
			fast = fast.next
		}
	}
}

func (se *stackWrapError) Unwrap() error {
	if se.next == nil {
		return se.error
	} else {
		return se.next.Unwrap()
	}
}

func (se *stackWrapError) UnwrapAll() []error {
	switch {
	case se.next != nil && se.error != nil:
		return []error{se.error, se.next}

	case se.next != nil:
		return []error{se.next}

	case se.error != nil:
		return []error{se.error}

	default:
		return nil
	}
}

func (se *stackWrapError) writeError(sb *strings.Builder) {
	sb.WriteString(se.Frame.String())

	if se.error != nil {
		sb.WriteString(": ")
		sb.WriteString(se.error.Error())
	}

	if se.next != nil {
		sb.WriteString("\n")
		se.next.writeError(sb)
	}

	if se.next == nil && se.error == nil {
		sb.WriteString("zit/alfa/errors/stackWrapError: both next and error are nil.")
		sb.WriteString("zit/alfa/errors/stackWrapError: this usually means that some nil error was wrapped in the error stack.")
	}
}

func (se *stackWrapError) writeErrorNoStack(sb *strings.Builder) {
	if se.ExtraData != "" {
		fmt.Fprintf(sb, "- %s\n", se.ExtraData)
	}

	if se.error != nil {
		fmt.Fprintf(sb, "- %s\n", se.error.Error())
	}

	if se.next != nil {
		se.next.writeErrorNoStack(sb)
	}

	if se.next == nil && se.error == nil {
		sb.WriteString("zit/alfa/errors/stackWrapError: both next and error are nil.")
		sb.WriteString("zit/alfa/errors/stackWrapError: this usually means that some nil error was wrapped in the error stack.")
	}
}

func (se *stackWrapError) Error() string {
	sb := &strings.Builder{}
	se.writeError(sb)
	return sb.String()
}
