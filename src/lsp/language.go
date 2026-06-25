package lsp

type LibMember struct {
	Params  []string
	Returns string
	Doc     string
	IsConst bool
}

type LibNamespace struct {
	Doc     string
	Members map[string]LibMember
}

type LanguageLibrary struct {
	Keywords      []LibMember
	ValidTypes    []string
	Builtins      []LibMember
	Namespaces    map[string]LibNamespace
	TypeRegistry  map[string]map[string]LibMember
}

func NewLanguageLibrary() *LanguageLibrary {
	lib := &LanguageLibrary{
		Keywords: []LibMember{
			{Params: nil, Returns: "", Doc: "Declares a mutable variable: `catch x: number = 0`", IsConst: false},
			{Params: nil, Returns: "", Doc: "Declares a constant: `anchor PI: number = 3.14159`. Must be assigned a value.", IsConst: false},
			{Params: nil, Returns: "", Doc: "Closes a block opened by `swim`, `if`, `while`, `for`, or `try`.", IsConst: false},
			{Params: nil, Returns: "", Doc: "Declares a function: `swim name(param: type): returnType ... shore`", IsConst: false},
			{Params: nil, Returns: "", Doc: "Returns a value from the enclosing function.", IsConst: false},
			{Params: nil, Returns: "", Doc: "Conditional statement.", IsConst: false},
			{Params: nil, Returns: "", Doc: "Alternate branch for `if`.", IsConst: false},
			{Params: nil, Returns: "", Doc: "Loops while a condition is truthy.", IsConst: false},
			{Params: nil, Returns: "", Doc: "Iterates: `for item in someArray ... shore`", IsConst: false},
			{Params: nil, Returns: "", Doc: "Used with `for` to iterate over an array's elements.", IsConst: false},
			{Params: nil, Returns: "", Doc: "Exits the nearest enclosing loop.", IsConst: false},
			{Params: nil, Returns: "", Doc: "Skips to the next iteration of the nearest enclosing loop.", IsConst: false},
			{Params: nil, Returns: "", Doc: "Begins an import statement: `from \"./file\" catch name, other as alias`", IsConst: false},
			{Params: nil, Returns: "", Doc: "Aliases an imported name: `from \"./mod\" catch thing as renamed`", IsConst: false},
			{Params: nil, Returns: "", Doc: "Marks a `swim`, `catch`, or `anchor` declaration as exported from the module.", IsConst: false},
			{Params: nil, Returns: "", Doc: "Returns the runtime type name of an expression as a string.", IsConst: false},
			{Params: nil, Returns: "", Doc: "Begins a try block: `try ... hook errName ... shore`", IsConst: false},
			{Params: nil, Returns: "", Doc: "Introduces the error-handling block of a `try` statement, binding the caught error to a name.", IsConst: false},
			{Params: nil, Returns: "", Doc: "Boolean literal.", IsConst: false},
			{Params: nil, Returns: "", Doc: "Boolean literal.", IsConst: false},
			{Params: nil, Returns: "", Doc: "The null literal.", IsConst: false},
			{Params: nil, Returns: "", Doc: "Logical AND.", IsConst: false},
			{Params: nil, Returns: "", Doc: "Logical OR.", IsConst: false},
		},
		ValidTypes: []string{"number", "string", "bool", "function", "void", "nil", "array", "object"},
		Builtins: []LibMember{
			{Params: []string{"...values"}, Returns: "void", Doc: "Prints all arguments to stdout, space-separated, followed by a newline."},
			{Params: []string{"value"}, Returns: "string", Doc: "Returns the runtime type of a value as a string (e.g. \"number\", \"array\")."},
			{Params: []string{"value"}, Returns: "number", Doc: "Converts a string or bool to a number. Strings must be valid numeric literals."},
			{Params: []string{"value"}, Returns: "string", Doc: "Converts any value to its string representation."},
			{Params: []string{"value"}, Returns: "number", Doc: "Returns the length of a string (rune count) or array."},
		},
		Namespaces: make(map[string]LibNamespace),
		TypeRegistry: make(map[string]map[string]LibMember),
	}

	lib.Namespaces["math"] = LibNamespace{
		Doc: "Built-in math functions and constants.",
		Members: map[string]LibMember{
			"floor":   {Params: []string{"n: number"}, Returns: "number", Doc: "Rounds n down to the nearest integer."},
			"ceil":    {Params: []string{"n: number"}, Returns: "number", Doc: "Rounds n up to the nearest integer."},
			"round":   {Params: []string{"n: number"}, Returns: "number", Doc: "Rounds n to the nearest integer."},
			"abs":     {Params: []string{"n: number"}, Returns: "number", Doc: "Returns the absolute value of n."},
			"min":     {Params: []string{"a: number", "b: number"}, Returns: "number", Doc: "Returns the smaller of a and b."},
			"max":     {Params: []string{"a: number", "b: number"}, Returns: "number", Doc: "Returns the larger of a and b."},
			"pow":     {Params: []string{"base: number", "exp: number"}, Returns: "number", Doc: "Returns base raised to the power of exp."},
			"sqrt":    {Params: []string{"n: number"}, Returns: "number", Doc: "Returns the square root of n."},
			"rand":    {Params: []string{}, Returns: "number", Doc: "Returns a random float in [0, 1)."},
			"randInt": {Params: []string{"min: number", "max: number"}, Returns: "number", Doc: "Returns a random integer in [min, max], inclusive."},
			"pi":      {Params: nil, Returns: "number", Doc: "The constant π (3.14159…).", IsConst: true},
			"e":       {Params: nil, Returns: "number", Doc: "Euler's number (2.71828…).", IsConst: true},
			"inf":     {Params: nil, Returns: "number", Doc: "Positive infinity.", IsConst: true},
		},
	}

	lib.Namespaces["string"] = LibNamespace{
		Doc: "Built-in string functions.",
		Members: map[string]LibMember{
			"upper":        {Params: []string{"s: string"}, Returns: "string", Doc: "Converts s to uppercase."},
			"lower":        {Params: []string{"s: string"}, Returns: "string", Doc: "Converts s to lowercase."},
			"trim":         {Params: []string{"s: string"}, Returns: "string", Doc: "Strips leading and trailing whitespace from s."},
			"split":        {Params: []string{"s: string", "sep: string"}, Returns: "[]string", Doc: "Splits s on sep, returning an array of parts."},
			"contains":     {Params: []string{"s: string", "sub: string"}, Returns: "bool", Doc: "Returns true if s contains sub."},
			"replace":      {Params: []string{"s: string", "old: string", "new: string"}, Returns: "string", Doc: "Replaces all occurrences of old with new in s."},
			"startsWith":   {Params: []string{"s: string", "prefix: string"}, Returns: "bool", Doc: "Returns true if s starts with prefix."},
			"endsWith":     {Params: []string{"s: string", "suffix: string"}, Returns: "bool", Doc: "Returns true if s ends with suffix."},
			"repeat":       {Params: []string{"s: string", "n: number"}, Returns: "string", Doc: "Returns s repeated n times."},
			"slice":        {Params: []string{"s: string", "start: number", "end: number"}, Returns: "string", Doc: "Returns the substring s[start:end)."},
			"indexOf":      {Params: []string{"s: string", "sub: string"}, Returns: "number", Doc: "Returns the index of the first occurrence of sub in s, or -1."},
			"charCode":     {Params: []string{"s: string"}, Returns: "number", Doc: "Returns the character code of the first rune in s."},
			"fromCharCode": {Params: []string{"code: number"}, Returns: "string", Doc: "Returns the single-character string for the given character code."},
		},
	}

	lib.Namespaces["array"] = LibNamespace{
		Doc: "Built-in array functions.",
		Members: map[string]LibMember{
			"push":     {Params: []string{"arr: array", "value"}, Returns: "array", Doc: "Returns a new array with value appended."},
			"pop":      {Params: []string{"arr: array"}, Returns: "value", Doc: "Returns the last element of arr."},
			"dropLast": {Params: []string{"arr: array"}, Returns: "array", Doc: "Returns a new array with the last element removed."},
			"sort":     {Params: []string{"arr: array"}, Returns: "array", Doc: "Returns a sorted copy of arr (all-number or all-string elements only)."},
			"reverse":  {Params: []string{"arr: array"}, Returns: "array", Doc: "Returns a reversed copy of arr."},
			"first":    {Params: []string{"arr: array"}, Returns: "value", Doc: "Returns the first element of arr."},
			"last":     {Params: []string{"arr: array"}, Returns: "value", Doc: "Returns the last element of arr."},
			"slice":    {Params: []string{"arr: array", "start: number", "end: number"}, Returns: "array", Doc: "Returns arr[start:end) as a new array."},
			"contains": {Params: []string{"arr: array", "value"}, Returns: "bool", Doc: "Returns true if arr contains value."},
			"join":     {Params: []string{"arr: array", "sep: string"}, Returns: "string", Doc: "Joins all elements of arr into a string separated by sep."},
		},
	}

	lib.Namespaces["tui"] = LibNamespace{
		Doc: "Terminal UI helpers: cursor control, color, progress bars, input.",
		Members: map[string]LibMember{
			"clear":   {Params: []string{}, Returns: "void", Doc: "Clears the terminal screen."},
			"move":    {Params: []string{"row: number", "col: number"}, Returns: "void", Doc: "Moves the cursor to the given terminal position."},
			"color":   {Params: []string{"color: string", "text: string"}, Returns: "string", Doc: "Wraps text in the given ANSI color."},
			"bar":     {Params: []string{"current: number", "max: number", "width: number"}, Returns: "string", Doc: "Returns a colored progress bar string."},
			"print":   {Params: []string{"...values"}, Returns: "void", Doc: "Prints to stdout without a trailing newline."},
			"println": {Params: []string{"...values"}, Returns: "void", Doc: "Prints to stdout with a trailing newline."},
			"input":   {Params: []string{"prompt: string"}, Returns: "string", Doc: "Reads a line of input from the user, showing prompt first."},
			"sleep":   {Params: []string{"ms: number"}, Returns: "void", Doc: "Sleeps for the given number of milliseconds."},
			"width":   {Params: []string{}, Returns: "number", Doc: "Returns the current terminal width in columns."},
			"height":  {Params: []string{}, Returns: "number", Doc: "Returns the current terminal height in rows."},
		},
	}

	lib.Namespaces["os"] = LibNamespace{
		Doc: "Filesystem, process, and script-environment helpers.",
		Members: map[string]LibMember{
			"read":      {Params: []string{"path: string"}, Returns: "string", Doc: "Reads the entire contents of the file at path as a string."},
			"write":     {Params: []string{"path: string", "contents: string"}, Returns: "void", Doc: "Writes contents to the file at path, overwriting any existing content."},
			"append":    {Params: []string{"path: string", "contents: string"}, Returns: "void", Doc: "Appends contents to the file at path, creating it if necessary."},
			"open":      {Params: []string{"path: string"}, Returns: "object", Doc: "Opens a file handle object at path for incremental writes."},
			"close":     {Params: []string{"handle: object"}, Returns: "void", Doc: "Closes a file handle previously returned by os.open."},
			"clock":     {Params: []string{}, Returns: "number", Doc: "Returns the current time, useful for timing and benchmarking."},
			"args":      {Params: []string{}, Returns: "[]string", Doc: "Returns the script's command-line arguments."},
			"scriptDir": {Params: nil, Returns: "string", Doc: "The absolute directory path of the currently running script.", IsConst: true},
		},
	}

	lib.Namespaces["json"] = LibNamespace{
		Doc: "JSON helpers.",
		Members: map[string]LibMember{
			"encode": {Params: []string{"obj: object"}, Returns: "string", Doc: "Encodes a TunaScript object to a JSON string."},
			"decode": {Params: []string{"json: string"}, Returns: "object", Doc: "Decodes a JSON string into a TunaScript object."},
		},
	}

	lib.Namespaces["imui"] = LibNamespace{
		Doc: "Immediate-mode GUI library for native desktop windows.",
		Members: map[string]LibMember{
			"createWindow": {Params: []string{"title: string", "width: number", "height: number", "tickFn: function"}, Returns: "void", Doc: "Creates a native window."},
			"button":       {Params: []string{"id: string", "text: string"}, Returns: "Button", Doc: "Draws a button and returns a persistent widget object."},
			"text":         {Params: []string{"id: string", "content: string"}, Returns: "Text", Doc: "Draws a text label and returns a persistent widget object."},
			"checkbox":     {Params: []string{"id: string", "label: string"}, Returns: "Checkbox", Doc: "Draws a checkbox and returns a persistent widget object."},
			"toggle":       {Params: []string{"id: string", "label: string"}, Returns: "Toggle", Doc: "Draws a toggle switch and returns a persistent widget object."},
			"slider":       {Params: []string{"id: string", "min: number", "max: number", "value: number"}, Returns: "Slider", Doc: "Draws a horizontal slider and returns a persistent widget object."},
			"frame":        {Params: []string{"id: string", "x: number", "y: number", "width: number", "height: number"}, Returns: "Frame", Doc: "Draws a bordered container and pushes a layout origin."},
			"endFrame":     {Params: []string{"id: string"}, Returns: "void", Doc: "Closes the frame opened by imui.frame(id)."},
			"setCursor":    {Params: []string{"type: string"}, Returns: "void", Doc: "Sets the OS mouse cursor."},
		},
	}

	commonMembers := map[string]LibMember{
		"px":            {Params: nil, Returns: "number", Doc: "Current X position (read-only).", IsConst: true},
		"py":            {Params: nil, Returns: "number", Doc: "Current Y position (read-only).", IsConst: true},
		"move":          {Params: []string{"x: number", "y: number"}, Returns: "void", Doc: "Moves the widget, overriding auto-layout."},
		"setSize":       {Params: []string{"width: number", "height: number"}, Returns: "void", Doc: "Sets the widget width and height in pixels."},
		"setAnchor":     {Params: []string{"x: number", "y: number"}, Returns: "void", Doc: "Sets the widget's anchor point."},
	}

	buttonMembers := copyMap(commonMembers)
	buttonMembers["clicked"] = LibMember{Params: nil, Returns: "bool", Doc: "True the frame this button was clicked (read-only).", IsConst: true}
	buttonMembers["onClick"] = LibMember{Params: []string{"fn: function"}, Returns: "void", Doc: "Called with no arguments on each click."}
	buttonMembers["onHover"] = LibMember{Params: []string{"fn: function"}, Returns: "void", Doc: "Called every frame while the mouse is over the button."}
	lib.TypeRegistry["Button"] = buttonMembers

	textMembers := copyMap(commonMembers)
	textMembers["text"] = LibMember{Params: nil, Returns: "string", Doc: "Override the displayed text string (assign).", IsConst: true}
	lib.TypeRegistry["Text"] = textMembers

	checkboxMembers := copyMap(commonMembers)
	checkboxMembers["checked"] = LibMember{Params: nil, Returns: "bool", Doc: "Current checked state (read-only).", IsConst: true}
	checkboxMembers["onChange"] = LibMember{Params: []string{"fn: function"}, Returns: "void", Doc: "Called with the new bool state each time it toggles."}
	lib.TypeRegistry["Checkbox"] = checkboxMembers

	toggleMembers := copyMap(commonMembers)
	toggleMembers["on"] = LibMember{Params: nil, Returns: "bool", Doc: "True when the toggle is on (read-only).", IsConst: true}
	toggleMembers["onChange"] = LibMember{Params: []string{"fn: function"}, Returns: "void", Doc: "Called with the new bool state each time it toggles."}
	lib.TypeRegistry["Toggle"] = toggleMembers

	sliderMembers := copyMap(commonMembers)
	sliderMembers["value"] = LibMember{Params: nil, Returns: "number", Doc: "Current slider value (read-only).", IsConst: true}
	sliderMembers["onChange"] = LibMember{Params: []string{"fn: function"}, Returns: "void", Doc: "Called with the new number value while dragging."}
	lib.TypeRegistry["Slider"] = sliderMembers

	frameMembers := copyMap(commonMembers)
	frameMembers["bgColor"] = LibMember{Params: nil, Returns: "string", Doc: "Background fill colour (assign).", IsConst: true}
	lib.TypeRegistry["Frame"] = frameMembers

	return lib
}

func copyMap(original map[string]LibMember) map[string]LibMember {
	result := make(map[string]LibMember, len(original))
	for k, v := range original {
		result[k] = v
	}
	return result
}
