package stack_frame

import (
	"fmt"
	"strings"
)

type Error struct {
	Err    error
	Frames []Frame
}

// TODO switch to having errors printed through a formatter

func (err Error) Error() string {
	var sb strings.Builder

	sb.WriteString("\n\nStack:\n")

	for _, frame := range err.Frames {
		fmt.Fprintln(&sb, frame.String())
	}

	sb.WriteString("\nError:\n")
	sb.WriteString(err.Err.Error())

	return sb.String()
}

func (err Error) Unwrap() error {
	return err.Err
}
