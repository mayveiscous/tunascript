package imui

type Renderer struct {
	fb		*Framebuffer
	clip		Rect
	clipStack	[]Rect
	textColor	uint32
}

func NewRenderer(fb *Framebuffer) *Renderer {
	return &Renderer{
		fb:	fb,
		clip:	Rect{0, 0, fb.Width, fb.Height},
	}
}

func (r *Renderer) Clear(color uint32) {
	r.fb.Clear(color)
}

func (r *Renderer) SetTextColor(c uint32) {
	r.textColor = c
}

func (r *Renderer) PushClip(rect Rect) {
	r.clipStack = append(r.clipStack, r.clip)
	r.clip = r.clip.Intersect(rect)
}

func (r *Renderer) PopClip() {
	if len(r.clipStack) == 0 {
		return
	}
	r.clip = r.clipStack[len(r.clipStack)-1]
	r.clipStack = r.clipStack[:len(r.clipStack)-1]
}

func (r *Renderer) SetPixel(x, y int, color uint32) {
	if x < r.clip.X || x >= r.clip.X+r.clip.W {
		return
	}
	if y < r.clip.Y || y >= r.clip.Y+r.clip.H {
		return
	}
	r.fb.Pixels[y*r.fb.Width+x] = color
}

func (r *Renderer) FillRect(x, y, w, h int, color uint32) {
	if w <= 0 || h <= 0 {
		return
	}
	cx := r.clip.X
	cy := r.clip.Y
	cw := r.clip.X + r.clip.W
	ch := r.clip.Y + r.clip.H

	sx := x
	if sx < cx {
		sx = cx
	}
	ex := x + w
	if ex > cw {
		ex = cw
	}
	sy := y
	if sy < cy {
		sy = cy
	}
	ey := y + h
	if ey > ch {
		ey = ch
	}
	if sx >= ex || sy >= ey {
		return
	}

	stride := r.fb.Width
	for row := sy; row < ey; row++ {
		idx := row*stride + sx
		for col := sx; col < ex; col++ {
			r.fb.Pixels[idx] = color
			idx++
		}
	}
}

func (r *Renderer) DrawRect(x, y, w, h int, color uint32, thickness int) {
	if thickness <= 0 {
		return
	}
	t := thickness
	r.FillRect(x, y, w, t, color)
	r.FillRect(x, y+h-t, w, t, color)
	r.FillRect(x, y, t, h, color)
	r.FillRect(x+w-t, y, t, h, color)
}

func (r *Renderer) DrawLine(x1, y1, x2, y2 int, color uint32) {
	dx := x2 - x1
	dy := y2 - y1
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	var sx, sy int
	if x2 > x1 {
		sx = 1
	} else {
		sx = -1
	}
	if y2 > y1 {
		sy = 1
	} else {
		sy = -1
	}
	err := dx - dy
	x, y := x1, y1
	for {
		r.SetPixel(x, y, color)
		if x == x2 && y == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

func (r *Renderer) FillRoundedRect(x, y, w, h, radius int, fill, border uint32, borderThk int) {
	if borderThk < 0 {
		borderThk = 0
	}
	if radius < 0 {
		radius = 0
	}
	if radius*2 > w {
		radius = w / 2
	}
	if radius*2 > h {
		radius = h / 2
	}

	innerX := x + radius
	innerY := y + radius
	innerW := w - 2*radius
	innerH := h - 2*radius

	r.FillRect(innerX, y, innerW, radius, fill)
	r.FillRect(innerX, y+h-radius, innerW, radius, fill)
	r.FillRect(x, innerY, w, innerH, fill)

	if radius > 0 && borderThk > 0 {
		r.DrawRect(x, y, w, h, border, borderThk)
		r.DrawArcCorners(x, y, w, h, radius, border, borderThk)
	} else if borderThk > 0 {
		r.DrawRect(x, y, w, h, border, borderThk)
	}
}

func (r *Renderer) DrawArcCorners(x, y, w, h, radius int, color uint32, thickness int) {
	if radius <= 0 || thickness <= 0 {
		return
	}
	_ = thickness
	for dy := 0; dy <= radius; dy++ {
		dd := radius*radius - dy*dy
		if dd < 0 {
			continue
		}
		dx := intSqrt(dd)
		r.SetPixel(x+radius-dx, y+radius-dy, color)
		r.SetPixel(x+radius+dx+(w-2*radius-1), y+radius-dy, color)
		r.SetPixel(x+radius-dx, y+radius+dy+(h-2*radius-1), color)
		r.SetPixel(x+radius+dx+(w-2*radius-1), y+radius+dy+(h-2*radius-1), color)
	}
}

func intSqrt(n int) int {
	if n <= 0 {
		return 0
	}
	x := n
	y := (x + 1) / 2
	for y < x {
		x = y
		y = (x + n/x) / 2
	}
	return x
}
