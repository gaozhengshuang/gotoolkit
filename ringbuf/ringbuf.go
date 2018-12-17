/// @file ringbuf.go
/// @brief 环形缓冲区
/// @author jackytse, xiejian1998@foxmail.com
/// @version 1.0
/// @date 2018-12-11

package ringbuf
//import "fmt"

// TODO: 当 wpos == rpos 要么size=0 或者size=cap
type Buffer struct {
	buf 	[]byte
	rpos 	int32
	wpos	int32
	size 	int32
	capacity int32
}


func NewBuffer(cap int32) *Buffer {
	rbuf := &Buffer{capacity:cap}
	rbuf.buf = make([]byte, rbuf.capacity, rbuf.capacity)
	rbuf.rpos = 0
	rbuf.wpos = 0
	rbuf.size = 0
	return rbuf
}


// --------------------------------------------------------------------------
/// @brief 导出接口
///
/// @param 
// --------------------------------------------------------------------------

func (r *Buffer) Len() int32 {
	return r.size
}

func (r *Buffer) Cap() int32 {
	return r.capacity
}

func (r *Buffer) Idle() int32 {
	return r.Cap() - r.Len()
}

func (r *Buffer) Reset() {
	r.rpos = 0
	r.wpos = 0
	r.size = 0
}

func (r *Buffer) Write(buf []byte) {
	wlen := int32(len(buf))
	if r.Idle() < wlen {
		r.growBuffer(wlen)
	}

	if r.wpos >= r.rpos {
		rightspace := r.Cap() - r.wpos
		if rightspace >= wlen {
			copy(r.buf[r.wpos:], buf[0:wlen])
			r.wpos += wlen
		}else {
			copy(r.buf[r.wpos:], buf[0:rightspace])
			copy(r.buf, buf[rightspace:])
			r.wpos = wlen - rightspace
		}
		r.size += wlen
	}else /*r.wpos < r.rpos*/ {
		copy(r.buf[r.wpos:], buf[0:wlen])
		r.wpos += wlen
		r.size += wlen
	}
}

// --------------------------------------------------------------------------
/// @brief Read 
/// @brief TODO: 为了效率，返回的切片某些情况下是Buffer内部切片的引用
///
/// @param num int32
// --------------------------------------------------------------------------
func (r *Buffer) Read(num int32) (buf []byte) {
	if num > r.Len() {
		return nil
	}

	if r.wpos <= r.rpos {
		rightspace := r.Cap() - r.rpos
		if rightspace >= num {
			buf = r.buf[r.rpos:r.rpos+num]	// buf is reference to r.buf
			r.rpos += num
		}else {
			buf = make([]byte, num, num)	// buf is a copy
			copy(buf, r.buf[r.rpos:])
			copy(buf[rightspace:], r.buf[0:num-rightspace])
			r.rpos = num-rightspace
		}

		r.size -= num
		return buf
	}else /*r.wpos > r.rpos*/ {
		buf := r.buf[r.rpos:r.rpos+num]		// buf is reference to r.buf
		r.rpos += num
		r.size -= num
		return buf
	}

	return nil
}

func (r *Buffer) View(num int32) (buf []byte) {
	if num > r.Len() {
		return nil
	}

	if r.wpos <= r.rpos {
		rightspace := r.Cap() - r.rpos
		if rightspace >= num {
			buf = r.buf[r.rpos:r.rpos+num]
		}else {
			buf = make([]byte, num, num)
			copy(buf, r.buf[r.rpos:])
			copy(buf[rightspace:], r.buf[0:num-rightspace])
		}
		return buf
	}else /*r.wpos > r.rpos*/ {
		buf := r.buf[r.rpos:r.rpos+num]
		return buf
	}

	return nil
}

// --------------------------------------------------------------------------
/// @brief 非导出借口
///
/// @param 
// --------------------------------------------------------------------------
func (r *Buffer) growBuffer(num int32) {
	if r.Idle() >= num {
		return
	}

	cap  := r.Cap()
	idle := cap - r.Len()
	for idle < num {
		cap += cap				// double rise
		idle = cap - r.Len()
	}

	newcap := cap
	buf := make([]byte, newcap, newcap)
	if r.wpos > r.rpos {
		copy(buf, r.buf[r.rpos:r.wpos])
	}else if r.wpos <= r.rpos {
		copy(buf, r.buf[r.rpos:r.Cap()])
		copy(buf[r.Cap()-r.rpos:], r.buf[0:r.wpos])
	}

	r.rpos = 0
	r.wpos = r.Len()
	r.capacity = newcap
	r.buf = buf
}


