package imui

type Rect struct {
	X, Y, W, H int
}

func (r Rect) Intersect(other Rect) Rect {
	nx := r.X
	if other.X > nx {
		nx = other.X
	}
	ny := r.Y
	if other.Y > ny {
		ny = other.Y
	}
	rw := r.X + r.W
	ow := other.X + other.W
	if ow < rw {
		rw = ow
	}
	rh := r.Y + r.H
	oh := other.Y + other.H
	if oh < rh {
		rh = oh
	}
	return Rect{X: nx, Y: ny, W: rw - nx, H: rh - ny}
}

func (r Rect) IsEmpty() bool	{ return r.W <= 0 || r.H <= 0 }
