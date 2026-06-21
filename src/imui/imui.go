package imui

import (
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

type Widget struct {
	ID	string
	Kind	string

	IdleColor	uint32
	HasIdleColor	bool
	HoverColor	uint32
	HasHoverColor	bool
	PressColor	uint32
	HasPressColor	bool
	Width		int
	HasWidth	bool
	Height		int
	HasHeight	bool
	OnClick		func()
	OnHover		func()

	OverrideText	string
	HasOverrideText	bool

	CheckColor	uint32
	HasCheckColor	bool
	OnChange	func(bool)

	OnColor		uint32
	HasOnColor	bool
	OffColor	uint32
	HasOffColor	bool
	KnobColor	uint32
	HasKnobColor	bool

	TrackColor	uint32
	HasTrackColor	bool
	HandleColor	uint32
	HasHandleColor	bool
	OverrideMin	float64
	HasOverrideMin	bool
	OverrideMax	float64
	HasOverrideMax	bool
	OnSlide		func(float64)

	TextColor		uint32
	HasTextColor		bool
	BorderColor		uint32
	HasBorderColor		bool
	BorderThickness		int
	HasBorderThickness	bool
	CornerRadius		int
	HasCornerRadius		bool
	Label			string
	HasLabel		bool

	BgColor			uint32
	HasBgColor		bool
	FrameBorderColor	uint32
	HasFrameBorderColor	bool
	Padding			int
	HasPadding		bool

	X, Y, W, H	int

	OverrideX, OverrideY	int
	HasOverrideXY		bool

	AnchorX, AnchorY	float64
	HasAnchor		bool
}

func (w *Widget) Move(x, y int) {
	w.OverrideX = x
	w.OverrideY = y
	w.HasOverrideXY = true
}

func (w *Widget) SetAnchor(nx, ny float64) {
	w.AnchorX = nx
	w.AnchorY = ny
	w.HasAnchor = true
}

func (w *Widget) SetSize(width, height int) {
	w.Width = width
	w.HasWidth = true
	w.Height = height
	w.HasHeight = true
	if w.Kind == "frame" {
		w.W = width
		w.H = height
	}
}

func (w *Widget) SetColor(field string, raw any) {
	c := parseColor(raw)
	switch field {
	case "idleColor":
		w.IdleColor = c
		w.HasIdleColor = true
	case "hoverColor":
		w.HoverColor = c
		w.HasHoverColor = true
	case "pressColor":
		w.PressColor = c
		w.HasPressColor = true
	case "textColor":
		w.TextColor = c
		w.HasTextColor = true
	case "checkColor":
		w.CheckColor = c
		w.HasCheckColor = true
	case "borderColor":
		w.BorderColor = c
		w.HasBorderColor = true
	case "onColor":
		w.OnColor = c
		w.HasOnColor = true
	case "offColor":
		w.OffColor = c
		w.HasOffColor = true
	case "knobColor":
		w.KnobColor = c
		w.HasKnobColor = true
	case "trackColor":
		w.TrackColor = c
		w.HasTrackColor = true
	case "handleColor":
		w.HandleColor = c
		w.HasHandleColor = true
	case "bgColor":
		w.BgColor = c
		w.HasBgColor = true
	case "frameBorderColor":
		w.FrameBorderColor = c
		w.HasFrameBorderColor = true
	}
}

func pick(val uint32, has bool, fallback uint32) uint32 {
	if has {
		return val
	}
	return fallback
}

var widgetRegistry = map[string]*Widget{}

func GetOrCreateWidget(id, kind string) *Widget {
	if w, ok := widgetRegistry[id]; ok {
		return w
	}
	w := &Widget{ID: id, Kind: kind}
	widgetRegistry[id] = w
	return w
}

type UIState struct {
	MouseX, MouseY	int
	MouseDown	bool
	ActiveID	string
	HoveredID	string
	CurrentIndex	int

	cursorStack		[]cursorEntry
	CursorX, CursorY	int

	DefaultStyle	Style
	ElementStyle	map[string]Style
	WidgetState	map[string]WidgetState

	RequestedCursor	string
}

type cursorEntry struct{ x, y int }

type Style struct {
	BgColor		uint32
	BtnIdleColor	uint32
	BtnHoverColor	uint32
	BtnPressColor	uint32
	TextColor	uint32
	Padding		int
}

type WidgetState struct {
	Bool	bool
	Float	float64
	Int	int
}

type ScriptCallback func()

var State UIState
var globalRenderer *Renderer

func InitStyle() {
	State.DefaultStyle = Style{
		BgColor:	0xFFF0F0F0,
		BtnIdleColor:	0xFFFFFFFF,
		BtnHoverColor:	0xFFE0E0E0,
		BtnPressColor:	0xFFAAAAAA,
		TextColor:	0xFF000000,
		Padding:	8,
	}
	State.ElementStyle = map[string]Style{}
	State.WidgetState = map[string]WidgetState{}
}

func ResetFrame() {
	State.CurrentIndex = 0
	State.CursorX = 20
	State.CursorY = 20
	State.cursorStack = nil
	State.RequestedCursor = ""
}

func getStyle(id string) Style {
	if s, ok := State.ElementStyle[id]; ok {
		return s
	}
	return State.DefaultStyle
}

func resolveStyle(id, fallback string) Style {
	if s, ok := State.ElementStyle[id]; ok {
		return s
	}
	if s, ok := State.ElementStyle[fallback]; ok {
		return s
	}
	return State.DefaultStyle
}

func getNextLayoutPosition(w, h int) (int, int) {
	x := State.CursorX
	y := State.CursorY
	padding := State.DefaultStyle.Padding
	if s, ok := State.ElementStyle["button"]; ok {
		padding = s.Padding
	}
	State.CursorY += h + padding
	return x, y
}

func pushLayout(x, y, w, h int, widget *Widget) {
	State.cursorStack = append(State.cursorStack, cursorEntry{State.CursorX, State.CursorY})
	pad := State.DefaultStyle.Padding
	if widget.HasPadding {
		pad = widget.Padding
	}
	State.CursorX = x + pad
	State.CursorY = y + pad
}

func popLayout(widget *Widget) {
	n := len(State.cursorStack)
	if n == 0 {
		return
	}
	restored := State.cursorStack[n-1]
	State.cursorStack = State.cursorStack[:n-1]
	State.CursorX = restored.x
	State.CursorY = widget.Y + widget.H + State.DefaultStyle.Padding
}

func parseColor(value any) uint32 {
	switch v := value.(type) {
	case uint32:
		return v
	case int:
		return uint32(v)
	case float64:
		return uint32(v)
	case string:
		s := strings.TrimSpace(v)
		s = strings.TrimPrefix(s, "#")
		s = strings.TrimPrefix(s, "0x")
		s = strings.TrimPrefix(s, "0X")
		if s == "" {
			return 0
		}
		parsed, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			return 0
		}
		return uint32(parsed)
	default:
		return 0
	}
}

func SetElement(id, field string, value any) {
	style := getStyle(id)
	color := parseColor(value)
	switch field {
	case "background":
		style.BgColor = color
	case "text":
		style.TextColor = color
	case "idle":
		style.BtnIdleColor = color
	case "hover":
		style.BtnHoverColor = color
	case "press":
		style.BtnPressColor = color
	}
	State.ElementStyle[id] = style
}

func SetColor(element string, value any) {
	color := parseColor(value)
	style := State.ElementStyle[element]
	if style == (Style{}) {
		style = State.DefaultStyle
	}
	switch element {
	case "background":
		style.BgColor = color
	case "button":
		style.BtnIdleColor = color
	case "text":
		style.TextColor = color
	}
	State.ElementStyle[element] = style
}

const (
	WS_OVERLAPPEDWINDOW	= 0x00CF0000
	WS_VISIBLE		= 0x10000000
	WM_DESTROY		= 0x0002
	WM_PAINT		= 0x000F
	WM_MOUSEMOVE		= 0x0200
	WM_LBUTTONDOWN		= 0x0201
	WM_LBUTTONUP		= 0x0202
	WM_SETCURSOR		= 0x0020
	HTCLIENT		= 1
	IDC_ARROW		= 32512
	IDC_HAND		= 32649
	IDC_IBEAM		= 32513
	IDC_CROSS		= 32515
	IDC_WAIT		= 32514
	IDC_SIZENS		= 32645
	IDC_SIZEWE		= 32644
	IDC_SIZEALL		= 32646
)

var (
	user32			= syscall.NewLazyDLL("user32.dll")
	gdi32			= syscall.NewLazyDLL("gdi32.dll")
	procDefWindowW		= user32.NewProc("DefWindowProcW")
	procRegisterW		= user32.NewProc("RegisterClassW")
	procCreateW		= user32.NewProc("CreateWindowExW")
	procGetMessage		= user32.NewProc("GetMessageW")
	procTranslate		= user32.NewProc("TranslateMessage")
	procDispatch		= user32.NewProc("DispatchMessageW")
	procBeginPaint		= user32.NewProc("BeginPaint")
	procEndPaint		= user32.NewProc("EndPaint")
	procInvalidate		= user32.NewProc("InvalidateRect")
	procGetClientRect	= user32.NewProc("GetClientRect")
	procDeleteObj		= gdi32.NewProc("DeleteObject")
	procSelectObj		= gdi32.NewProc("SelectObject")
	procCreateCompatibleDC	= gdi32.NewProc("CreateCompatibleDC")
	procDeleteDC		= gdi32.NewProc("DeleteDC")
	procBitBlt		= gdi32.NewProc("BitBlt")
	procCreateDIBSection	= gdi32.NewProc("CreateDIBSection")
	procLoadCursorW		= user32.NewProc("LoadCursorW")
	procSetCursor		= user32.NewProc("SetCursor")
)

type RECT struct{ Left, Top, Right, Bottom int32 }
type PAINTSTRUCT struct {
	Hdc			uintptr
	FErase			int32
	RcPaint			RECT
	FRestore, FIncUpdate	int32
	RgbReserved		[32]byte
}
type WNDCLASS struct {
	Style, LpfnWndProc				uintptr
	CbClsExtra, CbWndExtra				int32
	HInstance, HIcon, HCursor, HbrBackground	uintptr
	LpszMenuName, LpszClassName			*uint16
}
type MSG struct {
	Hwnd		uintptr
	Message		uint32
	WParam, LParam	uintptr
	Time		uint32
	Pt		int64
}

type BITMAPINFOHEADER struct {
	Size		uint32
	Width		int32
	Height		int32
	Planes		uint16
	BitCount	uint16
	Compression	uint32
	SizeImage	uint32
	XPelsPerMeter	int32
	YPelsPerMeter	int32
	ClrUsed		uint32
	ClrImportant	uint32
}

type BITMAPINFO struct {
	Header	BITMAPINFOHEADER
	Colors	[1]uint32
}

var (
	dibSurfaceFB	*Framebuffer
	dibDC		uintptr
	dibBmp		uintptr
	dibOldBmp	uintptr
	dibWidth	int
	dibHeight	int
)

func destroyDibSurface() {
	if dibSurfaceFB != nil {
		dibSurfaceFB.Pixels = nil
	}
	if dibOldBmp != 0 && dibDC != 0 {
		procSelectObj.Call(dibDC, dibOldBmp)
	}
	if dibBmp != 0 {
		procDeleteObj.Call(dibBmp)
	}
	if dibDC != 0 {
		procDeleteDC.Call(dibDC)
	}
	dibSurfaceFB = nil
	dibDC = 0
	dibBmp = 0
	dibOldBmp = 0
	dibWidth = 0
	dibHeight = 0
}

func CreateWindow(title string, width, height int, runScriptTick ScriptCallback) {
	InitStyle()
	className, _ := syscall.UTF16PtrFromString("SCRIPT_IMGUI")
	windowName, _ := syscall.UTF16PtrFromString(title)

	wndClass := WNDCLASS{
		LpfnWndProc: syscall.NewCallback(func(hwnd uintptr, msg uint32, wparam, lparam uintptr) uintptr {
			switch msg {
			case WM_MOUSEMOVE:
				newX := int(int16(lparam & 0xFFFF))
				newY := int(int16((lparam >> 16) & 0xFFFF))
				if newX != State.MouseX || newY != State.MouseY {
					State.MouseX = newX
					State.MouseY = newY

					procInvalidate.Call(hwnd, 0, 0)
				}
			case WM_LBUTTONDOWN:
				State.MouseDown = true
				procInvalidate.Call(hwnd, 0, 0)
			case WM_LBUTTONUP:
				State.MouseDown = false
				procInvalidate.Call(hwnd, 0, 0)
			case WM_PAINT:
				func() {
					var ps PAINTSTRUCT
					screenDC, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
					defer procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))

					var client RECT
					procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&client)))
					w := int(client.Right - client.Left)
					h := int(client.Bottom - client.Top)
					if w <= 0 || h <= 0 {
						return
					}

					if w != dibWidth || h != dibHeight {
						info := BITMAPINFO{Header: BITMAPINFOHEADER{
							Size:		uint32(unsafe.Sizeof(BITMAPINFOHEADER{})),
							Width:		int32(w),
							Height:		-int32(h),
							Planes:		1,
							BitCount:	32,
						}}
						var pBits unsafe.Pointer
						newBmp, _, _ := procCreateDIBSection.Call(
							screenDC, uintptr(unsafe.Pointer(&info)), 0,
							uintptr(unsafe.Pointer(&pBits)), 0, 0)
						if newBmp == 0 {
							return
						}
						newDC, _, _ := procCreateCompatibleDC.Call(screenDC)
						if newDC == 0 {
							procDeleteObj.Call(newBmp)
							return
						}
						newOldBmp := selectObject(newDC, newBmp)
						if newOldBmp == 0 {
							procDeleteDC.Call(newDC)
							procDeleteObj.Call(newBmp)
							return
						}

						destroyDibSurface()

						dibBmp, dibDC, dibOldBmp = newBmp, newDC, newOldBmp
						dibWidth, dibHeight = w, h

						if dibSurfaceFB == nil {
							dibSurfaceFB = NewFramebuffer(w, h)
						}
						pixelSlice := unsafe.Slice((*uint32)(pBits), w*h)
						dibSurfaceFB.WrapExternal(pixelSlice, w, h)
					}

					renderer := NewRenderer(dibSurfaceFB)
					bg := State.DefaultStyle
					if s, ok := State.ElementStyle["background"]; ok {
						bg = s
					}
					renderer.Clear(bg.BgColor)
					globalRenderer = renderer

					func() {
						defer func() { recover() }()
						ResetFrame()
						runScriptTick()
					}()

					if State.RequestedCursor == "" {
						State.RequestedCursor = "arrow"
					}

					procBitBlt.Call(screenDC, 0, 0, uintptr(w), uintptr(h),
						dibDC, 0, 0, 0x00CC0020)
				}()
				return 0
			case WM_SETCURSOR:
				if lparam&0xFFFF == HTCLIENT {
					curs := State.RequestedCursor
					if curs == "" {
						curs = "arrow"
					}
					if h := loadCursor(curs); h != 0 {
						procSetCursor.Call(h)
						return 1
					}
				}
			case WM_DESTROY:
				destroyDibSurface()
				syscall.ExitProcess(0)
			}
			ret, _, _ := procDefWindowW.Call(hwnd, uintptr(msg), wparam, lparam)
			return ret
		}),
		HbrBackground:	0,
		LpszClassName:	className,
	}
	procRegisterW.Call(uintptr(unsafe.Pointer(&wndClass)))
	procCreateW.Call(
		0, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(windowName)),
		WS_OVERLAPPEDWINDOW|WS_VISIBLE, 100, 100, uintptr(width), uintptr(height), 0, 0, 0, 0,
	)
	var msg MSG
	for {
		ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		procTranslate.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatch.Call(uintptr(unsafe.Pointer(&msg)))
	}
	destroyDibSurface()
}

func selectObject(hdc uintptr, obj uintptr) uintptr {
	ret, _, _ := procSelectObj.Call(hdc, obj)
	return ret
}

var cursorCache = map[string]uintptr{}

func cursorNameToID(name string) uint32 {
	switch name {
	case "arrow":
		return IDC_ARROW
	case "hand":
		return IDC_HAND
	case "text":
		return IDC_IBEAM
	case "cross":
		return IDC_CROSS
	case "wait":
		return IDC_WAIT
	case "resizeNS":
		return IDC_SIZENS
	case "resizeWE":
		return IDC_SIZEWE
	case "resizeAll":
		return IDC_SIZEALL
	default:
		return IDC_ARROW
	}
}

func loadCursor(typeStr string) uintptr {
	if h, ok := cursorCache[typeStr]; ok {
		return h
	}
	id := cursorNameToID(typeStr)
	if id == 0 {
		return 0
	}
	h, _, _ := procLoadCursorW.Call(0, uintptr(id))
	if h != 0 {
		cursorCache[typeStr] = h
	}
	return h
}

func SetCursor(cursorType string) {
	State.RequestedCursor = cursorType
}
