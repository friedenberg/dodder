package stack_frame

func MakeErrorTree(
	err error,
	frames ...Frame,
) *ErrorTree {
	var tree *ErrorTree

	if tree, _ = err.(*ErrorTree); tree == nil {
		tree = &ErrorTree{
			Root:        err,
			Descendents: make(ErrorsAndFrames, 0, len(frames)),
		}
	}

	// switch err := err.(type) {
	// case *ErrorTree:
	// case ErrorsAndFramesGetter:
	// 	tree.Descendents.AppendExisting(err.GetErrorsAndFrames()...)
	// }

	for idx := range frames {
		tree.Descendents.Append(nil, frames[idx])
	}

	return tree
}

type ErrorTree struct {
	Root        error
	Descendents ErrorsAndFrames
}

func (tree *ErrorTree) Error() string {
	if tree.Root != nil {
		return tree.Root.Error()
	} else if len(tree.Descendents) > 0 {
		return "error tree"
	} else {
		return "empty error tree"
	}
}

func (tree *ErrorTree) Unwrap() error {
	return tree.Root
}

func (tree *ErrorTree) Append(childError error, childFrame Frame) {
	tree.Descendents.Append(childError, childFrame)
}

func (tree *ErrorTree) GetErrorsAndFrames() []ErrorAndFrame {
	return tree.Descendents
}
