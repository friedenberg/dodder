package organize_text

type ErrorRead struct {
	error

	line, column int
}
