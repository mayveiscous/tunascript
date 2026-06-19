# Tunascript

A small scripting language with an ocean-themed syntax, implemented in Go. Zero runtime dependencies, static analysis before execution.

Licensed under MIT.

---

## Installation

Go to https://tunascript.org and install the latest version of **tunascript**.
Add the directory you installed it to, to your path and restart your **IDE** if necessary.

---

## Quick Example

```tuna
swim greet(name: string): string
    serve "Hello, " + name + "!"
shore

catch names = ["Tuna", "Salmon", "Mackeral"]
for name in names
    bubble(greet(name))
shore
```

---

## Syntax Reference

### Comments

```
><> Single-line comment

></
Multi-line
comment
/>
```

### Variables

```
><> Mutable
catch x: number = 10
catch y = "hello"
catch z as 42

><> Constant (must be assigned)
anchor MAX: number = 100
anchor NAME = "Tuna"
```

### Functions

```
><> Named function
swim add(a: number, b: number): number
    serve a + b
shore

><> Anonymous function expression
catch double = swim(n: number): number
    serve n * 2
shore

><> Inline invocation
bubble((swim(x: number, y: number): number
    serve x * y
shore)(3, 4))

><> Variadic parameters
swim sum(...nums: number): number
    ...

><> Exported (for modules)
cast swim greet(name: string): string
    serve "Hi " + name
shore
```

### Conditionals

```
if score >= 90
    bubble("A")
else if score >= 70
    bubble("C")
else
    bubble("F")
shore
```

### Loops

```
><> While
while i < 5
    bubble(i)
    i++
shore

><> For-in (array)
for item in items
    bubble(item)
shore

><> For-in (object keys and values)
for key, val in point
    bubble(key + " = " + toString(val))
shore

break      ><> Exit the loop
continue   ><> Skip to next iteration
```

### Error Handling

```
try
    catch data = os.read("file.txt")
    bubble(data)
hook err
    bubble("Error: " + err)
shore
```

### Types / Structs

```
school Point = {
    x: number,
    y: number
}

><> Usage
catch p: Point = { x: 10, y: 20 }
```

### Operators

| Category | Operators |
|----------|-----------|
| Arithmetic | `+` `-` `*` `/` `%` |
| Comparison | `==` `!=` `<` `>` `<=` `>=` |
| Logical | `and` `or` `!` |
| Compound | `+=` `-=` `*=` `/=` `++` `--` |
| Other | `=` (assign), `.` (member), `[]` (index), `()` (call), `typeof` |
| Swap | `a, b = c, d` (parallel assignment) |

### Modules

```
><> utils.tuna — export declarations
cast swim square(n: number): number
    serve n * n
shore

cast anchor TAX_RATE = 0.08

><> main.tuna — import with optional aliases
from "utils.tuna" catch square, TAX_RATE as tax

bubble(square(5))  ><< 25
bubble(tax)        ><< 0.08
```

### Types

`number`, `string`, `bool`, `null`, `function`, `array`, `object`, `void`

---

## Standard Library

### Globals

| Function | Returns | Description |
|---|---|---|
| `bubble(...values)` | void | Prints args, space-separated, with newline |
| `len(value)` | number | Length of a string or array |
| `typeOf(value)` | string | Runtime type name as string |
| `toString(value)` | string | Converts any value to string |
| `toNumber(value)` | number | Converts string/bool to number |

### `math`

`floor` `ceil` `round` `abs` `min` `max` `pow` `sqrt` `rand` `randInt` `pi` `e` `inf`

### `string`

`upper` `lower` `trim` `split` `contains` `replace` `startsWith` `endsWith` `repeat` `slice` `indexOf` `charCode` `fromCharCode`

### `array`

`push` `pop` `dropLast` `sort` `reverse` `first` `last` `slice` `contains` `join`

### `tui`

| Function | Description |
|---|---|
| `tui.clear()` | Clears the terminal |
| `tui.move(row, col)` | Moves cursor to position |
| `tui.print(...values)` | Prints without newline |
| `tui.println(...values)` | Prints with newline |
| `tui.input(prompt)` | Reads a line of input |
| `tui.color(color, text)` | Wraps text in ANSI color (`red` `green` `yellow` `blue` `cyan` `magenta` `white` `bold` `dim`) |
| `tui.bar(current, max, width)` | Renders a `█░` progress bar string |
| `tui.sleep(ms)` | Pauses for `ms` milliseconds |
| `tui.width()` | Terminal width in columns |
| `tui.height()` | Terminal height in rows |

### `os`

| Function | Returns | Description |
|---|---|---|
| `os.read(path)` | string | Reads entire file as string |
| `os.write(path, contents)` | void | Writes string to file (overwrites) |
| `os.append(path, contents)` | void | Appends string to file |
| `os.open(path)` | object | Opens file handle for incremental writes |
| `os.close(handle)` | void | Closes file handle |
| `os.clock()` | number | Current time in milliseconds |
| `os.args()` | array | Command-line arguments |

### `json`

| Function | Returns | Description |
|---|---|---|
| `json.encode(obj)` | string | Encodes object to JSON string |
| `json.decode(str)` | object | Decodes JSON string to object |

### `imui` (Win32 only)

Immediate-mode GUI for native Windows desktop windows.

| Function | Description |
|---|---|
| `imui.createWindow(title, width, height, tickFn)` | Creates a native window. `tickFn` is called every frame. |
| `imui.frame(id, x, y, w, h)` | Draws a bordered container and pushes a layout origin |
| `imui.endFrame(id)` | Closes a frame |
| `imui.button(id, text)` | Draws a clickable button |
| `imui.text(id, content)` | Draws a label |
| `imui.checkbox(id, label)` | Draws a toggleable checkbox |
| `imui.toggle(id, label)` | Draws a toggle switch |
| `imui.slider(id, min, max, value)` | Draws a draggable slider |

All widgets return persistent objects with assignable properties (`.px`, `.py`, `.clicked`, `.textColor`, `.borderColor`, etc.) and methods (`.onClick(fn)`, `.move(x, y)`, `.onChange(fn)`).

---

## Static Analysis

Before execution, the interpreter runs a static analysis pass that checks for errors and warnings:

| Code | Level | Detects |
|------|-------|---------|
| E001 | error | `serve` outside a function |
| E002 | error | `break` outside a loop |
| E003 | error | `continue` outside a loop |
| E004 | error | `break`/`continue` crossing a function boundary |
| E005 | error | Undefined variable |
| E006 | error | Assignment to constant |
| E007 | error | `break` inside function but no enclosing loop |
| W001 | warning | Unused variable or parameter |
| W002 | warning | Declaration overwrites a builtin name |
| W003 | warning | Variable shadows an outer declaration |
| W004 | warning | Assignment overwrites a builtin |
| W005 | warning | Unreachable code after `serve`/`break`/`continue` |

Errors prevent execution; warnings and hints are printed but allow execution.