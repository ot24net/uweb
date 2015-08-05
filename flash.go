package uweb

//
// Depends on session
//
func MdFlash() Middleware {
	return new(Flashing)
}

//
// Flashing
//
type Flashing struct {
	// empty
}

// Impl Middleware
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

func (f *Flash) key(k string) string {
	return "_flash_" + k
}

// Put flash msg for use on next request
func (f *Flash) Put(k, v string) {
	if len(k) == 0 {
		return
	}
	k = f.key(k)
	f.sess.Set(k, v)
	f.sess.Save()
}

// Pop flash msg and release, only once.
func (f *Flash) Pop(k string) string {
	if len(k) == 0 {
		return ""
	}
	k = f.key(k)
	v := f.sess.Get(k)
	if len(v) != 0 {
		f.sess.Del(k)
		f.sess.Save()
	}
	return v
}
