package uweb

const (
	// break from middleare stacks and
	// will not write response by Response.End
	NEXT_ABORT = -1

	// break from middleware stacks
	NEXT_BREAK = 0

	// continue in middleware stacks
	NEXT_CONTINUE = 1
)

//
// Middleware is for code-level extending
// but user-level extending, as the golang
// is not as dynamic as js or ruby.
//
type Middleware interface {
	Handle(*Context) int
}
