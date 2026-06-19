package imui

func applyAnchor(widget *Widget, x, y, w, h int) (int, int) {
	if widget.HasAnchor {
		x = x - int(widget.AnchorX*float64(w))
		y = y - int(widget.AnchorY*float64(h))
	}
	return x, y
}

func Text(id string, text string) *Widget {
	widget := GetOrCreateWidget(id, "text")

	display := text
	if widget.HasOverrideText { display = widget.OverrideText }

	w, h := len(display)*8, 16
	var x, y int
	if widget.HasOverrideXY {
		x, y = widget.OverrideX, widget.OverrideY
	} else {
		x, y = getNextLayoutPosition(w, h)
	}
	widget.X, widget.Y = x, y
	x, y = applyAnchor(widget, x, y, w, h)

	base := resolveStyle(id, "text")
	tc := pick(widget.TextColor, widget.HasTextColor, base.TextColor)

	procSetTextColor.Call(globalHDC, uintptr(tc))
	procSetBkMode.Call(globalHDC, 1)
	drawRawText(globalHDC, display, x, y)

	return widget
}

func Button(id string, text string) (*Widget, bool) {
	State.CurrentIndex++
	widget := GetOrCreateWidget(id, "button")

	base := resolveStyle(id, "button")
	w := 120; if widget.HasWidth  { w = widget.Width  }
	h := 30;  if widget.HasHeight { h = widget.Height }

	var x, y int
	if widget.HasOverrideXY {
		x, y = widget.OverrideX, widget.OverrideY
	} else {
		x, y = getNextLayoutPosition(w, h)
	}
	widget.X, widget.Y = x, y
	x, y = applyAnchor(widget, x, y, w, h)

	isHovered := State.MouseX >= x && State.MouseX <= x+w &&
		State.MouseY >= y && State.MouseY <= y+h

	if isHovered {
		State.HoveredID = id
		if State.MouseDown && State.ActiveID == "" {
			State.ActiveID = id
		}
		if widget.OnHover != nil { widget.OnHover() }
	}

	idleC  := pick(widget.IdleColor,  widget.HasIdleColor,  base.BtnIdleColor)
	hoverC := pick(widget.HoverColor, widget.HasHoverColor, base.BtnHoverColor)
	pressC := pick(widget.PressColor, widget.HasPressColor, base.BtnPressColor)
	textC  := pick(widget.TextColor,  widget.HasTextColor,  base.TextColor)

	var color uint32
	switch {
	case State.ActiveID == id && isHovered: color = pressC
	case isHovered:                          color = hoverC
	default:                                 color = idleC
	}

	borderC   := pick(widget.BorderColor,      widget.HasBorderColor,      0x00000000)
	borderThk := 1
	if widget.HasBorderThickness { borderThk = widget.BorderThickness }
	cornerR   := 0
	if widget.HasCornerRadius    { cornerR = widget.CornerRadius }

	drawFilledRect(globalHDC, x, y, w, h, color, borderC, borderThk, cornerR)

	procSetTextColor.Call(globalHDC, uintptr(textC))
	procSetBkMode.Call(globalHDC, 1)
	textW := len(text) * 8
	textH := 16
	tx := x + (w-textW)/2
	if tx < x {
		tx = x
	}
	ty := y + (h-textH)/2
	if ty < y {
		ty = y
	}
	if tx+textW <= x+w && ty+textH <= y+h {
		drawRawText(globalHDC, text, tx, ty)
	}

	clicked := false
	if !State.MouseDown && State.ActiveID == id {
		State.ActiveID = ""
		if isHovered {
			clicked = true
			if widget.OnClick != nil { widget.OnClick() }
		}
	}

	return widget, clicked
}

func Checkbox(id string, label string) (*Widget, bool) {
	State.CurrentIndex++
	widget := GetOrCreateWidget(id, "checkbox")

	displayLabel := label
	if widget.HasLabel { displayLabel = widget.Label }

	size := 18
	if widget.HasWidth { size = widget.Width }

	var x, y int
	if widget.HasOverrideXY {
		x, y = widget.OverrideX, widget.OverrideY
	} else {
		x, y = getNextLayoutPosition(size, size)
	}
	widget.X, widget.Y = x, y
	x, y = applyAnchor(widget, x, y, size, size)

	ws := State.WidgetState[id]
	hovered := State.MouseX >= x && State.MouseX <= x+size &&
		State.MouseY >= y && State.MouseY <= y+size

	wasChecked := ws.Bool
	if hovered && State.MouseDown && State.ActiveID == "" {
		State.ActiveID = id
	}
	if !State.MouseDown && State.ActiveID == id && hovered {
		ws.Bool = !ws.Bool
		State.WidgetState[id] = ws
		State.ActiveID = ""
		if widget.OnChange != nil { widget.OnChange(ws.Bool) }
	}

	_ = wasChecked

	borderC   := pick(widget.BorderColor, widget.HasBorderColor, 0x00000000)
	checkC    := pick(widget.CheckColor,  widget.HasCheckColor,  0x0000AA00)
	textC     := pick(widget.TextColor,   widget.HasTextColor,   State.DefaultStyle.TextColor)
	borderThk := 1
	if widget.HasBorderThickness { borderThk = widget.BorderThickness }
	cornerR   := 0
	if widget.HasCornerRadius    { cornerR = widget.CornerRadius }

	boxColor := uint32(0x00FFFFFF)
	if ws.Bool { boxColor = checkC }

	drawFilledRect(globalHDC, x, y, size, size, boxColor, borderC, borderThk, cornerR)

	procSetTextColor.Call(globalHDC, uintptr(textC))
	procSetBkMode.Call(globalHDC, 1)
	drawRawText(globalHDC, displayLabel, x+size+6, y)

	return widget, ws.Bool
}

func Toggle(id string, label string) (*Widget, bool) {
	State.CurrentIndex++
	widget := GetOrCreateWidget(id, "toggle")

	displayLabel := label
	if widget.HasLabel { displayLabel = widget.Label }

	trackW, trackH := 48, 22
	if widget.HasWidth  { trackW = widget.Width }
	if widget.HasHeight { trackH = widget.Height }

	var x, y int
	if widget.HasOverrideXY {
		x, y = widget.OverrideX, widget.OverrideY
	} else {
		x, y = getNextLayoutPosition(trackW+80, trackH)
	}
	widget.X, widget.Y = x, y
	x, y = applyAnchor(widget, x, y, trackW, trackH)

	ws := State.WidgetState[id]
	hovered := State.MouseX >= x && State.MouseX <= x+trackW &&
		State.MouseY >= y && State.MouseY <= y+trackH

	if hovered && State.MouseDown && State.ActiveID == "" {
		State.ActiveID = id
	}
	if !State.MouseDown && State.ActiveID == id && hovered {
		ws.Bool = !ws.Bool
		State.WidgetState[id] = ws
		State.ActiveID = ""
		if widget.OnChange != nil { widget.OnChange(ws.Bool) }
	}

	onC   := pick(widget.OnColor,   widget.HasOnColor,   0x0000CC66)
	offC  := pick(widget.OffColor,  widget.HasOffColor,  0x00444444)
	knobC := pick(widget.KnobColor, widget.HasKnobColor, 0x00FFFFFF)
	textC := pick(widget.TextColor, widget.HasTextColor, State.DefaultStyle.TextColor)

	trackColor := offC
	if ws.Bool { trackColor = onC }
	if hovered { trackColor += 0x00111111 }

	trackBrush := createSolidBrush(trackColor)
	trackRect := RECT{int32(x), int32(y), int32(x+trackW), int32(y+trackH)}
	fillRect(globalHDC, &trackRect, trackBrush)
	deleteObject(trackBrush)

	knobX := x + 2
	if ws.Bool { knobX = x + trackW - trackH + 2 }
	knobBrush := createSolidBrush(knobC)
	knobRect := RECT{int32(knobX), int32(y+2), int32(knobX+trackH-4), int32(y+trackH-2)}
	fillRect(globalHDC, &knobRect, knobBrush)
	deleteObject(knobBrush)

	procSetTextColor.Call(globalHDC, uintptr(textC))
	procSetBkMode.Call(globalHDC, 1)
	drawRawText(globalHDC, displayLabel, x+trackW+8, y+4)

	return widget, ws.Bool
}

func Slider(id string, min, max, currentValue float64) (*Widget, float64) {
	State.CurrentIndex++
	widget := GetOrCreateWidget(id, "slider")

	if widget.HasOverrideMin { min = widget.OverrideMin }
	if widget.HasOverrideMax { max = widget.OverrideMax }

	trackW, trackH := 160, 8
	if widget.HasWidth  { trackW = widget.Width }
	if widget.HasHeight { trackH = widget.Height }

	var x, y int
	if widget.HasOverrideXY {
		x, y = widget.OverrideX, widget.OverrideY
	} else {
		x, y = getNextLayoutPosition(trackW, trackH+8)
	}
	widget.X, widget.Y = x, y
	x, y = applyAnchor(widget, x, y, trackW, trackH)
	ty := y + 4

	ws := State.WidgetState[id]
	if !widget.HasOverrideMin && ws.Float == 0 { ws.Float = min }

	hovered := State.MouseX >= x && State.MouseX <= x+trackW &&
		State.MouseY >= ty-6 && State.MouseY <= ty+trackH+6

	if hovered && State.MouseDown {
		State.ActiveID = id
		t := float64(State.MouseX-x) / float64(trackW)
		if t < 0 { t = 0 }
		if t > 1 { t = 1 }
		prev := ws.Float
		ws.Float = min + (max-min)*t
		State.WidgetState[id] = ws
		if ws.Float != prev && widget.OnSlide != nil { widget.OnSlide(ws.Float) }
	}
	if !State.MouseDown && State.ActiveID == id { State.ActiveID = "" }

	trackC  := pick(widget.TrackColor,  widget.HasTrackColor,  0x00CCCCCC)
	handleC := pick(widget.HandleColor, widget.HasHandleColor, 0x00555555)

	trackBrush := createSolidBrush(trackC)
	trackRect := RECT{int32(x), int32(ty), int32(x+trackW), int32(ty+trackH)}
	fillRect(globalHDC, &trackRect, trackBrush)
	deleteObject(trackBrush)

	t := 0.0
	if max > min { t = (ws.Float - min) / (max - min) }
	hx := x + int(float64(trackW)*t)
	handleBrush := createSolidBrush(handleC)
	hr := RECT{int32(hx-5), int32(ty-4), int32(hx+5), int32(ty+trackH+4)}
	fillRect(globalHDC, &hr, handleBrush)
	deleteObject(handleBrush)

	return widget, ws.Float
}


func Frame(id string, x, y, w, h int) *Widget {
	State.CurrentIndex++
	widget := GetOrCreateWidget(id, "frame")

	if widget.HasOverrideXY {
		x, y = widget.OverrideX, widget.OverrideY
	}
	widget.X, widget.Y, widget.W, widget.H = x, y, w, h
	fx, fy := applyAnchor(widget, x, y, w, h)

	bgC      := pick(widget.BgColor,            widget.HasBgColor,           0x00FAFAFA)
	borderC  := pick(widget.FrameBorderColor,   widget.HasFrameBorderColor,  0x00AAAAAA)
	// Also check the shared BorderColor as fallback
	if !widget.HasFrameBorderColor && widget.HasBorderColor {
		borderC = widget.BorderColor
	}
	borderThk := 1
	if widget.HasBorderThickness { borderThk = widget.BorderThickness }
	cornerR   := 0
	if widget.HasCornerRadius    { cornerR = widget.CornerRadius }

	drawFilledRect(globalHDC, fx, fy, w, h, bgC, borderC, borderThk, cornerR)

	pushLayout(fx, fy, w, h, widget)
	return widget
}

func EndFrame(id string) {
	widget, ok := widgetRegistry[id]
	if !ok {
		n := len(State.cursorStack)
		if n > 0 { State.cursorStack = State.cursorStack[:n-1] }
		return
	}
	popLayout(widget)
}