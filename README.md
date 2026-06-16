# Tunascript

A small scripting language built in Go. Ocean-themed syntax, no dependencies, surprisingly readable.

Licensed under MIT.

---

## Installation

1. Download the latest release and copy the path to the executable, e.g. `C:/Program Files/Tunascript/bin/`
2. Add that path to your `Path` under **Edit Environment Variables → User Variables**
3. Open a new terminal and run any `.tuna` file:

```bash
tuna run file.tuna
```

or 

```bash
tuna file.tuna
```

---

## Syntax

```tuna
><> this is a comment

><> variables and constants
catch name = "tuna"
catch age: number = 3
anchor MAX: number = 100

><> functions — closed with 'shore', return with 'serve'
swim add(a: number, b: number): number
  serve a + b
shore

bubble(add(3, 7))   ><> 10

><> conditionals
if score >= 90
  bubble("A")
else if score >= 70
  bubble("C")
else
  bubble("F")
shore

><> while loop
catch i = 0
while i < 5
  bubble(i)
  i++
shore

><> for-in loop
catch fish = ["salmon", "mackerel", "tuna"]
for f in fish
  bubble(f)
shore

><> objects
catch point = { x: 10, y: 20 }
point.x = 99
```

---

## Standard Library

### Globals
| Function | Returns | Description |
|---|---|---|
| `bubble(...values)` | void | Prints all arguments separated by spaces |
| `len(value)` | number | Length of a string or array |
| `typeOf(value)` | string | Type name of a value |
| `toString(value)` | string | Converts any value to a string |
| `toNumber(value)` | number | Converts a string or bool to a number |

### `math`
`floor` `ceil` `round` `abs` `min` `max` `pow` `sqrt` `rand` `randInt` `pi` `e` `inf`

### `string`
`upper` `lower` `trim` `split` `contains` `replace` `startsWith` `endsWith` `repeat` `indexOf`

### `array`
`push` `pop` `dropLast` `sort` `reverse` `first` `last` `slice` `contains` `join`

### `tui`
| Function | Description |
|---|---|
| `tui.clear()` | Clears the terminal |
| `tui.move(row, col)` | Moves the cursor |
| `tui.print(...values)` | Prints without a newline |
| `tui.println(...values)` | Prints with a newline |
| `tui.input(prompt)` | Reads a line of user input |
| `tui.color(color, text)` | Wraps text in an ANSI color (`red` `green` `yellow` `blue` `cyan` `magenta` `white` `bold` `dim`) |
| `tui.bar(current, max, width)` | Renders a `█░` progress bar string |
| `tui.sleep(ms)` | Pauses execution for `ms` milliseconds |
| `tui.width()` | Terminal width in columns |
| `tui.height()` | Terminal height in rows |

---

## Modules

Use `cast` to export, `from`/`catch` to import.

```tuna
><> utils.tuna
cast swim square(n: number): number
  serve n * n
shore

cast anchor TAX_RATE = 0.08
```

```tuna
><> main.tuna
from "utils.tuna" catch square, TAX_RATE as tax

bubble(square(5))  ><> 25
bubble(tax)        ><> 0.08
```