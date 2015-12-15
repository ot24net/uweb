package uweb

//
// Path Ignore middleware
//
func MdIgnore(ps []string) Middleware {
	ig := NewIgnore()
	for _, p := range ps {
		ig.Path(p)
	}
	return ig
}

//
// Ignore some path
//
type Ignore struct {
	paths map[string]bool
}

// Need call Path after create
func NewIgnore() *Ignore {
	return &Ignore{
		paths: make(map[string]bool),
	}
}

// Add one path
func (ig *Ignore) Path(p string) {
	ig.paths[p] = true
}

func (ig *Ignore) Name() string {
	return "ignore"
}

// Check request path, if ignored, return 200 ok and ignored text info
// @impl Middleware
func (ig *Ignore) Handle(c *Context) int {
	if _, ok := ig.paths[c.Req.URL.Path]; ok {
		c.Res.Status = 200
		c.Res.Body = []byte("Ignore: path is ignored")
		return NEXT_BREAK
	}
	return NEXT_CONTINUE
}
