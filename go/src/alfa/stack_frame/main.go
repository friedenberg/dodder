package stack_frame

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	cwd          string
	maxCallDepth int
)

func init() {
	var err error

	if cwd, err = os.Getwd(); err != nil {
		log.Panic(err)
	}
}

type Frame struct {
	Package     string
	Function    string
	Filename    string
	RelFilename string
	Line        int

	Prefix string

	nonZero bool
}

func MakeFrameFromRuntimeFrame(runtimeFrame runtime.Frame) (frame Frame) {
	frame.Filename = filepath.Clean(runtimeFrame.File)
	frame.Line = runtimeFrame.Line
	frame.Function = runtimeFrame.Function
	frame.Package, frame.Function = getPackageAndFunctionName(frame.Function)

	frame.RelFilename, _ = filepath.Rel(cwd, frame.Filename)
	frame.nonZero = true

	return
}

//go:noinline
func MustFrame(skip int) Frame {
	frame, ok := MakeFrame(skip + 1)

	if !ok {
		panic("stack unavailable")
	}

	return frame
}

//go:noinline
func MakeFrames(skip, count int) (frames []Frame) {
	programCounters := make([]uintptr, count)
	writtenCounters := runtime.Callers(skip+1, programCounters) // 0 is self
	if writtenCounters == 0 {
		return
	}

	programCounters = programCounters[:writtenCounters]

	rawFrames := runtime.CallersFrames(programCounters)

	frames = make([]Frame, 0, len(programCounters))

	for {
		frame, more := rawFrames.Next()
		frames = append(frames, MakeFrameFromRuntimeFrame(frame))

		if !more {
			break
		}
	}

	return
}

//go:noinline
func MakeFrame(skip int) (si Frame, ok bool) {
	var programCounter uintptr
	programCounter, _, _, ok = runtime.Caller(skip + 1) // 0 is self

	if !ok {
		return
	}

	frames := runtime.CallersFrames([]uintptr{programCounter})

	frame, _ := frames.Next()
	si = MakeFrameFromRuntimeFrame(frame)

	// TODO remove this ugly hack
	if si.Function == "Wrap" {
		panic(fmt.Sprintf("Parent Wrap included in stack. Skip: %d", skip))
	}

	return
}

func getPackageAndFunctionName(v string) (p string, f string) {
	p, f = filepath.Split(v)

	idx := strings.Index(f, ".")

	if idx == -1 {
		return
	}

	p += f[:idx]

	if len(f) > idx+1 {
		f = f[idx+1:]
	}

	return
}

func (si Frame) FileNameLine() string {
	filename := si.Filename

	if si.RelFilename != "" {
		filename = si.RelFilename
	}

	return fmt.Sprintf(
		"%s:%d",
		filename,
		si.Line,
	)
}

func (frame Frame) String() string {
	filename := frame.Filename

	if frame.RelFilename != "" {
		filename = frame.RelFilename
	}

	// TODO-P3 determine if si.line is ever not valid
	return fmt.Sprintf(
		"# %s\n%s%s:%d",
		frame.Function,
		frame.Prefix,
		filename,
		frame.Line,
	)
}

func (frame Frame) StringLogLine() string {
	filename := frame.Filename

	if frame.RelFilename != "" {
		filename = frame.RelFilename
	}

	// TODO-P3 determine if si.line is ever not valid
	return fmt.Sprintf(
		"%s%s:%d: %s",
		frame.Prefix,
		filename,
		frame.Line,
		frame.Function,
	)
}

func (si Frame) StringNoFunctionName() string {
	filename := si.Filename

	if si.RelFilename != "" {
		filename = si.RelFilename
	}

	return fmt.Sprintf(
		"%s|%d|",
		filename,
		si.Line,
	)
}

// If the frame is non-zero, return a wrapped error. Otherwise return the input
// error unwrapped.
func (frame Frame) Wrap(in error) (err error) {
	if frame.nonZero {
		return &stackWrapError{
			Frame: frame,
			error: in,
		}
	} else {
		return in
	}
}

func (si Frame) Wrapf(in error, f string, values ...any) (err error) {
	return &stackWrapError{
		Frame:     si,
		ExtraData: fmt.Sprintf(f, values...),
		next: &stackWrapError{
			Frame: si,
			error: in,
		},
	}
}

func (si Frame) Errorf(f string, values ...any) (err error) {
	return &stackWrapError{
		Frame: si,
		error: fmt.Errorf(f, values...),
	}
}

type stackWrapError struct {
	ExtraData string
	Frame
	error

	next *stackWrapError
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
		sb.WriteString("zit/alfa/stack_frame/stackWrapError: both next and error are nil.")
		sb.WriteString("zit/alfa/stack_frame/stackWrapError: this usually means that some nil error was wrapped in the error stack.")
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
		sb.WriteString("zit/alfa/stack_frame/stackWrapError: both next and error are nil.")
		sb.WriteString("zit/alfa/stack_frame/stackWrapError: this usually means that some nil error was wrapped in the error stack.")
	}
}

func (se *stackWrapError) Error() string {
	sb := &strings.Builder{}
	se.writeError(sb)
	return sb.String()
}
