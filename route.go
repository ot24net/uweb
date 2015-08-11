package uweb

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

//
// export as middleware
//
func MdRouter() Middleware {
	return defaultRouter
}

//
// Default router, for convienent
//
var (
	defaultRouter = NewRouter()
)

// Get handle
func Get(p string, h HttpHandler) {
	defaultRouter.Get(p, h)
}

// Post handle
func Post(p string, h HttpHandler) {
	defaultRouter.Post(p, h)
}

//
// Handler is handler for http request
//
type HttpHandler func(c *Context)

//
// Tree node
//
type RNode struct {
	child   []*RNode    // children
	height  int         // tree height, for fast match
	pattern string      // path pattern
	handler HttpHandler // only last height has h
}

// Dump internal status
func (n *RNode) Dump(indent string) {
	// dump self
	if len(indent) == 0 {
		indent = " "
	}
	fmt.Printf("%s pattern:%s, height:%d, handler:%d, child:%d\n", indent+indent, n.pattern, n.height, n.handler, len(n.child))

	// dump child
	for _, c := range n.child {
		c.Dump(indent + indent)
	}
}

// Add child node
func (n *RNode) Add(ps []string, handler HttpHandler) (int, error) {
	ps = append([]string{n.pattern}, ps...)
	if ok, err := n.merge(ps, handler); err != nil {
		return 0, err
	} else if ok {
		n.calc()
	}
	return n.height, nil
}

var (
	ErrDupPath = errors.New("RNode: dup path")
)

// Merge path to node
func (n *RNode) merge(ps []string, handler HttpHandler) (bool, error) {
	// check ps
	if len(ps) == 0 {
		return false, nil
	}

	// match current node and check dup
	if n.pattern != ps[0] {
		return false, nil
	}
	if len(ps) == 1 {
		if n.height == 1 || n.handler != nil {
			return false, ErrDupPath
		}
		n.handler = handler
		return true, nil
	}

	// let child merge first
	ps = ps[1:]
	merged := false
	for _, c := range n.child {
		if ok, err := c.merge(ps, handler); err != nil {
			return false, err
		} else if ok {
			merged = true
			break
		}
	}

	// if child not merged
	if !merged {
		nodes := make([]*RNode, len(ps))
		for i, p := range ps {
			nodes[i] = &RNode{
				pattern: p,
			}
			if i > 0 {
				parent := nodes[i-1]
				parent.child = append(parent.child, nodes[i])
			}
		}
		nodes[len(nodes)-1].handler = handler // only last node owns handler
		n.child = append(n.child, nodes[0])
	}

	// ok
	return true, nil
}

// Calc calcuate height of every node
func (n *RNode) calc() int {
	max := 0
	for _, c := range n.child {
		h := c.calc()
		if max < h {
			max = h
		}
	}
	n.height = max + 1
	return n.height
}

// Match patten with path array, and return matched pattern
func (n *RNode) Match(ps []string, ms map[string]string) *RNode {
	// if height not equal, ignore
	s := len(ps)
	if n.height < s {
		return nil
	}
	if s == 0 {
		return nil
	}

	// if pattern match fail, ignore
	p0 := ps[0]
	if n.pattern[0:1] != ":" && n.pattern != p0 {
		return nil
	}

	// if current node matched
	if len(ps) == 1 {
		if n.height == 1 {
			if n.pattern[0:1] == ":" {
				ms[n.pattern[1:]] = p0
			}
			return n
		} else if n.pattern == p0 {
			return n
		}
		return nil
	}

	// match child first
	for _, c := range n.child {
		if h := c.Match(ps[1:], ms); h != nil {
			if n.pattern[0:1] == ":" {
				ms[n.pattern[1:]] = p0
			}
			return h
		}
	}

	// fail
	return nil
}

//
// RTree is path router tree, for fast match
//
type RTree struct {
	mu   sync.Mutex
	root *RNode
}

// Create a tree with a root node with patten "/"
func NewRTree() *RTree {
	root := &RNode{
		pattern: "/",
	}
	return &RTree{
		root: root,
	}
}

// convert to path array
func (rt *RTree) parsePath(p string) []string {
	return strings.Split(strings.Trim(p, "/"), "/")
}

// Add path to tree
func (rt *RTree) Add(p string, h HttpHandler) error {
	ps := rt.parsePath(p)

	rt.mu.Lock()
	defer rt.mu.Unlock()

	if _, err := rt.root.Add(ps, h); err != nil {
		return err
	}
	return nil
}

// Match path and get handler
func (rt *RTree) Match(p string) (map[string]string, HttpHandler) {
	ps := append([]string{"/"}, rt.parsePath(p)...)
	ms := make(map[string]string)

	rt.mu.Lock()
	defer rt.mu.Unlock()

	if n := rt.root.Match(ps, ms); n != nil {
		return ms, n.handler
	}
	return nil, nil
}

//
// Router is a restfull path router
//
type Router struct {
	gets  *RTree
	puts  *RTree
	posts *RTree
	dels  *RTree
}

// Create default router
func NewRouter() *Router {
	return &Router{
		gets:  NewRTree(),
		puts:  NewRTree(),
		posts: NewRTree(),
		dels:  NewRTree(),
	}
}

// Middleware impl
func (r *Router) Handle(c *Context) int {
	// get tree by method
	var t *RTree
	switch c.Req.Method {
	case "GET":
		t = r.gets
	case "POST":
		t = r.posts
	case "PUT":
		t = r.puts
	case "DELETE":
		t = r.dels
	default:
		c.Res.Status = 501
		c.Res.Err = errors.New("Router: method not support yet")
		return NEXT_BREAK
	}

	// then match
	p, h := t.Match(c.Req.URL.Path)
	if h == nil {
		c.Res.Status = 404
		c.Res.Err = errors.New("Router: not found")
		return NEXT_BREAK
	}

	// handle
	c.Req.Params = p
	h(c)
	return NEXT_CONTINUE
}

// add handler to method trees
func (r *Router) addHandler(method, p string, h HttpHandler) {
	// method
	var t *RTree
	switch method {
	case "GET":
		t = r.gets
	case "POST":
		t = r.posts
	case "PUT":
		t = r.puts
	case "DELETE":
		t = r.dels
	default:
		panic("Router: method not support yet")
	}

	// add
	if err := t.Add(p, h); err != nil {
		panic(err)
	}
}

// Get
func (r *Router) Get(p string, h HttpHandler) {
	r.addHandler("GET", p, h)
}

// Post
func (r *Router) Post(p string, h HttpHandler) {
	r.addHandler("POST", p, h)
}

// Put
func (r *Router) Put(p string, h HttpHandler) {
	r.addHandler("PUT", p, h)
}

// Delete
func (r *Router) Del(p string, h HttpHandler) {
	r.addHandler("DELETE", p, h)
}
