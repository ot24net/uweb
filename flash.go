package uweb

//
// Create flash middleware,
// which depends on session middleware
//
func MdFlash() Middleware {
	return new(Flashing)
}

//
// Flashing middleware
//
type Flashing struct {
	// empty
}

// @impl Middleware
func (f *Flashing) Handle(c *Context) int {
	c.Flash = &Flash{c.Sess}
	return NEXT_CONTINUE
}

//
// Per request flash
//
type Flash struct {
	sess *Session
}

// add flash prefix
func (f *Flash) key(k string) string {
	return "_flash_" + k
}

// Put flash msg for Pop in next request
func (f *Flash) Put(k, v string) {
	k = f.key(k)
	f.sess.Set(k, v)
}

// Pop flash msg and release it.
func (f *Flash) Pop(k string) string {
	k = f.key(k)
	v := f.sess.Get(k)
	if len(v) != 0 {
		f.sess.Del(k)
	}
	return v
}
