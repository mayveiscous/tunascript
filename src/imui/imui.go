package imui

import (
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

// ---------------------------------------------------------------------------
// Widget — persistent handle for every widget, lives across frames
// ---------------------------------------------------------------------------

type Widget struct {
	ID   string
	Kind string // "button" | "text" | "checkbox" | "toggle" | "slider" | "frame"

	// ---- Button ----
	IdleColor  uint32; HasIdleColor  bool
	HoverColor uint32; HasHoverColor bool
	PressColor uint32; HasPressColor bool
	Width      int;    HasWidth      bool
	Height     int;    HasHeight     bool
	OnClick    func()
	OnHover    func()

	// ---- Text ----
	// TextColor shared below; OverrideText lets script set .text property
	OverrideText    string; HasOverrideText bool

	// ---- Checkbox ----
	CheckColor  uint32; HasCheckColor  bool
	OnChange    func(bool)   // checkbox, toggle

	// ---- Toggle ----
	OnColor   uint32; HasOnColor   bool
	OffColor  uint32; HasOffColor  bool
	KnobColor uint32; HasKnobColor bool

	// ---- Slider ----
	TrackColor  uint32; HasTrackColor  bool
	HandleColor uint32; HasHandleColor bool
	OverrideMin float64; HasOverrideMin bool
	OverrideMax float64; HasOverrideMax bool
	OnSlide     func(float64)

	// ---- Shared ----
	TextColor      uint32; HasTextColor      bool
	BorderColor    uint32; HasBorderColor    bool
	BorderThickness int;   HasBorderThickness bool
	CornerRadius   int;    HasCornerRadius   bool
	Label          string; HasLabel          bool

	// ---- Frame ----
	BgColor    uint32; HasBgColor    bool
	FrameBorderColor uint32; HasFrameBorderColor bool
	Padding    int;    HasPadding    bool

	// Frame geometry (stored so EndFrame can restore cursor correctly)
	X, Y, W, H int

	// ---- Override positioning (set by .move()) ----
	OverrideX, OverrideY int
	HasOverrideXY        bool

	// ---- Anchor point (0-1 normalized, set by .setAnchor()) ----
	// (0,0) = top-left (default), (0.5,0.5) = center, (1,1) = bottom-right
	AnchorX, AnchorY float64
	HasAnchor        bool
}

// Move sets an override position, bypassing auto-layout next frame.
func (w *Widget) Move(x, y int) {
	w.OverrideX = x
	w.OverrideY = y
	w.HasOverrideXY = true
}

// SetAnchor sets the anchor point for positioning.
// (0,0) = top-left corner, (0.5,0.5) = center, (1,1) = bottom-right.
func (w *Widget) SetAnchor(nx, ny float64) {
	w.AnchorX = nx
	w.AnchorY = ny
	w.HasAnchor = true
}

// SetSize overrides the widget width and height.
// For frames this also updates the W/H geometry fields.
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

// color helper — parse and mark a named color field
func (w *Widget) SetColor(field string, raw any) {
	c := parseColor(raw)
	switch field {
	case "idleColor":        w.IdleColor = c;        w.HasIdleColor = true
	case "hoverColor":       w.HoverColor = c;       w.HasHoverColor = true
	case "pressColor":       w.PressColor = c;       w.HasPressColor = true
	case "textColor":        w.TextColor = c;        w.HasTextColor = true
	case "checkColor":       w.CheckColor = c;       w.HasCheckColor = true
	case "borderColor":      w.BorderColor = c;      w.HasBorderColor = true
	case "onColor":          w.OnColor = c;          w.HasOnColor = true
	case "offColor":         w.OffColor = c;         w.HasOffColor = true
	case "knobColor":        w.KnobColor = c;        w.HasKnobColor = true
	case "trackColor":       w.TrackColor = c;       w.HasTrackColor = true
	case "handleColor":      w.HandleColor = c;      w.HasHandleColor = true
	case "bgColor":          w.BgColor = c;          w.HasBgColor = true
	case "frameBorderColor": w.FrameBorderColor = c; w.HasFrameBorderColor = true
	}
}

// or(a, hasA, b) — return a if explicitly set, else b
func pick(val uint32, has bool, fallback uint32) uint32 {
	if has { return val }
	return fallback
}

// ---------------------------------------------------------------------------
// Widget registry — persistent across frames
// ---------------------------------------------------------------------------

var widgetRegistry = map[string]*Widget{}

func GetOrCreateWidget(id, kind string) *Widget {
	if w, ok := widgetRegistry[id]; ok {
		return w
	}
	w := &Widget{ID: id, Kind: kind}
	widgetRegistry[id] = w
	return w
}

// ---------------------------------------------------------------------------
// UI State
// ---------------------------------------------------------------------------

type UIState struct {
	MouseX, MouseY int
	MouseDown      bool
	ActiveID       string
	HoveredID      string
	CurrentIndex   int

	cursorStack []cursorEntry
	CursorX, CursorY int

	DefaultStyle Style
	ElementStyle map[string]Style
	WidgetState  map[string]WidgetState
}

type cursorEntry struct{ x, y int }

type Style struct {
	BgColor       uint32
	BtnIdleColor  uint32
	BtnHoverColor uint32
	BtnPressColor uint32
	TextColor     uint32
	Padding       int
}

type WidgetState struct {
	Bool  bool
	Float float64
	Int   int
}

type ScriptCallback func(hdc uintptr)

var State UIState
var globalHDC uintptr

func InitStyle() {
	State.DefaultStyle = Style{
		BgColor:       0x00F0F0F0,
		BtnIdleColor:  0x00FFFFFF,
		BtnHoverColor: 0x00E0E0E0,
		BtnPressColor: 0x00AAAAAA,
		TextColor:     0x00000000,
		Padding:       8,
	}
	State.ElementStyle = map[string]Style{}
	State.WidgetState  = map[string]WidgetState{}
}

func ResetFrame(hdc uintptr) {
	globalHDC = hdc
	State.CurrentIndex = 0
	State.CursorX = 20
	State.CursorY = 20
	State.cursorStack = nil
}

func getStyle(id string) Style {
	if s, ok := State.ElementStyle[id]; ok { return s }
	return State.DefaultStyle
}

func resolveStyle(id, fallback string) Style {
	if s, ok := State.ElementStyle[id]; ok { return s }
	if s, ok := State.ElementStyle[fallback]; ok { return s }
	return State.DefaultStyle
}

func getNextLayoutPosition(w, h int) (int, int) {
	x := State.CursorX
	y := State.CursorY
	padding := State.DefaultStyle.Padding
	if s, ok := State.ElementStyle["button"]; ok { padding = s.Padding }
	State.CursorY += h + padding
	return x, y
}

// ---------------------------------------------------------------------------
// Frame layout stack
// ---------------------------------------------------------------------------

func pushLayout(x, y, w, h int, widget *Widget) {
	State.cursorStack = append(State.cursorStack, cursorEntry{State.CursorX, State.CursorY})
	pad := State.DefaultStyle.Padding
	if widget.HasPadding { pad = widget.Padding }
	State.CursorX = x + pad
	State.CursorY = y + pad
}

func popLayout(widget *Widget) {
	n := len(State.cursorStack)
	if n == 0 { return }
	restored := State.cursorStack[n-1]
	State.cursorStack = State.cursorStack[:n-1]
	State.CursorX = restored.x
	State.CursorY = widget.Y + widget.H + State.DefaultStyle.Padding
}

// ---------------------------------------------------------------------------
// Color parsing
// ---------------------------------------------------------------------------

func parseColor(value any) uint32 {
	switch v := value.(type) {
	case uint32:   return v
	case int:      return uint32(v)
	case float64:  return uint32(v)
	case string:
		s := strings.TrimSpace(v)
		s = strings.TrimPrefix(s, "#")
		s = strings.TrimPrefix(s, "0x")
		s = strings.TrimPrefix(s, "0X")
		if s == "" { return 0 }
		parsed, err := strconv.ParseUint(s, 16, 32)
		if err != nil { return 0 }
		return uint32(parsed)
	default:
		return 0
	}
}

func SetElement(id, field string, value any) {
	style := getStyle(id)
	color := parseColor(value)
	switch field {
	case "background": style.BgColor      = color
	case "text":       style.TextColor     = color
	case "idle":       style.BtnIdleColor  = color
	case "hover":      style.BtnHoverColor = color
	case "press":      style.BtnPressColor = color
	}
	State.ElementStyle[id] = style
}

func SetColor(element string, value any) {
	color := parseColor(value)
	style := State.ElementStyle[element]
	if style == (Style{}) { style = State.DefaultStyle }
	switch element {
	case "background": style.BgColor     = color
	case "button":     style.BtnIdleColor = color
	case "text":       style.TextColor    = color
	}
	State.ElementStyle[element] = style
}

// ---------------------------------------------------------------------------
// Win32
// ---------------------------------------------------------------------------

const (
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_VISIBLE          = 0x10000000
	WM_DESTROY          = 0x0002
	WM_PAINT            = 0x000F
	WM_MOUSEMOVE        = 0x0200
	WM_LBUTTONDOWN      = 0x0201
	WM_LBUTTONUP        = 0x0202
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	gdi32            = syscall.NewLazyDLL("gdi32.dll")
	procDefWindowW   = user32.NewProc("DefWindowProcW")
	procRegisterW    = user32.NewProc("RegisterClassW")
	procCreateW      = user32.NewProc("CreateWindowExW")
	procGetMessage   = user32.NewProc("GetMessageW")
	procTranslate    = user32.NewProc("TranslateMessage")
	procDispatch     = user32.NewProc("DispatchMessageW")
	procBeginPaint   = user32.NewProc("BeginPaint")
	procEndPaint     = user32.NewProc("EndPaint")
	procFillRect     = user32.NewProc("FillRect")
	procFrameRect    = user32.NewProc("FrameRect")
	procInvalidate   = user32.NewProc("InvalidateRect")
	procTextOutW     = gdi32.NewProc("TextOutW")
	procBrush        = gdi32.NewProc("CreateSolidBrush")
	procDeleteObj    = gdi32.NewProc("DeleteObject")
	procSetTextColor = gdi32.NewProc("SetTextColor")
	procSetBkMode    = gdi32.NewProc("SetBkMode")
	procCreatePen    = gdi32.NewProc("CreatePen")
	procSelectObj    = gdi32.NewProc("SelectObject")
	procRoundRect = gdi32.NewProc("RoundRect")
	procRectangle = gdi32.NewProc("Rectangle")
)

type RECT struct{ Left, Top, Right, Bottom int32 }
type PAINTSTRUCT struct {
	Hdc                  uintptr
	FErase               int32
	RcPaint              RECT
	FRestore, FIncUpdate int32
	RgbReserved          [32]byte
}
type WNDCLASS struct {
	Style, LpfnWndProc                              uintptr
	CbClsExtra, CbWndExtra                          int32
	HInstance, HIcon, HCursor, HbrBackground        uintptr
	LpszMenuName, LpszClassName                     *uint16
}
type MSG struct {
	Hwnd           uintptr
	Message        uint32
	WParam, LParam uintptr
	Time           uint32
	Pt             int64
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
					procInvalidate.Call(hwnd, 0, 1)
				}
			case WM_LBUTTONDOWN:
				State.MouseDown = true
				procInvalidate.Call(hwnd, 0, 1)
			case WM_LBUTTONUP:
				State.MouseDown = false
				procInvalidate.Call(hwnd, 0, 1)
			case WM_PAINT:
				var ps PAINTSTRUCT
				hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
				bg := State.DefaultStyle
				if s, ok := State.ElementStyle["background"]; ok { bg = s }
				bgBrush := createSolidBrush(bg.BgColor)
				procFillRect.Call(hdc, uintptr(unsafe.Pointer(&ps.RcPaint)), bgBrush)
				deleteObject(bgBrush)
				ResetFrame(hdc)
				runScriptTick(hdc)
				procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
				return 0
			case WM_DESTROY:
				syscall.ExitProcess(0)
			}
			ret, _, _ := procDefWindowW.Call(hwnd, uintptr(msg), wparam, lparam)
			return ret
		}),
		HbrBackground: 0,
		LpszClassName: className,
	}
	procRegisterW.Call(uintptr(unsafe.Pointer(&wndClass)))
	procCreateW.Call(
		0, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(windowName)),
		WS_OVERLAPPEDWINDOW|WS_VISIBLE, 100, 100, uintptr(width), uintptr(height), 0, 0, 0, 0,
	)
	var msg MSG
	for {
		ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(ret) <= 0 { break }
		procTranslate.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatch.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func createSolidBrush(color uint32) uintptr { ret, _, _ := procBrush.Call(uintptr(color)); return ret }
func deleteObject(obj uintptr)              { procDeleteObj.Call(obj) }
func fillRect(hdc uintptr, rc *RECT, br uintptr)  { procFillRect.Call(hdc, uintptr(unsafe.Pointer(rc)), br) }
func frameRect(hdc uintptr, rc *RECT, br uintptr) { procFrameRect.Call(hdc, uintptr(unsafe.Pointer(rc)), br) }
func createPen(style, width int, color uint32) uintptr {
	ret, _, _ := procCreatePen.Call(uintptr(style), uintptr(width), uintptr(color))
	return ret
}
func selectObject(hdc uintptr, obj uintptr) uintptr {
	ret, _, _ := procSelectObj.Call(hdc, obj)
	return ret
}

// drawFilledRect draws a filled rect with optional border and rounded corners.
// Uses pen-based GDI when borderThickness>1 or cornerRadius>0; otherwise falls
// back to the simpler FillRect+FrameRect approach.
func drawFilledRect(hdc uintptr, x, y, w, h int, fillColor, borderColor uint32, borderThickness, cornerRadius int) {
	if cornerRadius > 0 || borderThickness > 1 {
		pen := createPen(4, borderThickness, borderColor) // PS_INSIDEFRAME
		brush := createSolidBrush(fillColor)
		oldPen := selectObject(hdc, pen)
		oldBrush := selectObject(hdc, brush)
		if cornerRadius > 0 {
			procRoundRect.Call(hdc, uintptr(x), uintptr(y), uintptr(x+w), uintptr(y+h),
				uintptr(cornerRadius), uintptr(cornerRadius))
		} else {
			procRectangle.Call(hdc, uintptr(x), uintptr(y), uintptr(x+w), uintptr(y+h))
		}
		selectObject(hdc, oldPen)
		selectObject(hdc, oldBrush)
		deleteObject(pen)
		deleteObject(brush)
	} else if borderThickness <= 0 {
		rc := RECT{int32(x), int32(y), int32(x+w), int32(y+h)}
		fillBrush := createSolidBrush(fillColor)
		fillRect(hdc, &rc, fillBrush)
		deleteObject(fillBrush)
	} else {
		rc := RECT{int32(x), int32(y), int32(x+w), int32(y+h)}
		fillBrush := createSolidBrush(fillColor)
		fillRect(hdc, &rc, fillBrush)
		deleteObject(fillBrush)
		if borderColor != 0 {
			borderBrush := createSolidBrush(borderColor)
			frameRect(hdc, &rc, borderBrush)
			deleteObject(borderBrush)
		}
	}
}
func drawRawText(hdc uintptr, txt string, x, y int) {
	ptr, _ := syscall.UTF16PtrFromString(txt)
	procTextOutW.Call(hdc, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(ptr)), uintptr(len(txt)))
}