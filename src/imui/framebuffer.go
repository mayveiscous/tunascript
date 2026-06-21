package imui

type Framebuffer struct {
	Width	int
	Height	int
	Pixels	[]uint32
}

func NewFramebuffer(w, h int) *Framebuffer {
	return &Framebuffer{
		Width:	w,
		Height:	h,
		Pixels:	make([]uint32, w*h),
	}
}

func (fb *Framebuffer) Resize(w, h int) {
	if fb.Width == w && fb.Height == h {
		return
	}
	fb.Width = w
	fb.Height = h
	fb.Pixels = make([]uint32, w*h)
}

func (fb *Framebuffer) Clear(color uint32) {
	for i := range fb.Pixels {
		fb.Pixels[i] = color
	}
}

func (fb *Framebuffer) WrapExternal(pixels []uint32, w, h int) {
	fb.Pixels = pixels
	fb.Width = w
	fb.Height = h
}
