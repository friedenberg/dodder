// Code generated by "stringer -type=ContextState"; DO NOT EDIT.

package interfaces

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ContextStateUnknown-0]
	_ = x[ContextStateUnstarted-1]
	_ = x[ContextStateStarted-2]
	_ = x[ContextStateSucceeded-3]
	_ = x[ContextStateFailed-4]
	_ = x[ContextStateAborted-5]
}

const _ContextState_name = "ContextStateUnknownContextStateUnstartedContextStateStartedContextStateSucceededContextStateFailedContextStateAborted"

var _ContextState_index = [...]uint8{0, 19, 40, 59, 80, 98, 117}

func (i ContextState) String() string {
	if i >= ContextState(len(_ContextState_index)-1) {
		return "ContextState(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ContextState_name[_ContextState_index[i]:_ContextState_index[i+1]]
}
