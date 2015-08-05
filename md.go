package uweb

const (
	// abort without end
	NEXT_ABORT = -1

	// break middleware stacks
	NEXT_BREAK = 0

	// continue middleware stacks
	NEXT_CONTINUE = 1
)

//
// Middleware is for code-level extending
// not for user-level extending, as the golang
// is not as dynamic as js or ruby.
//
type Middleware interface {
	// return NEXT_BREAK or NEXT_CONTINUE
	Handle(*Context) int
}
