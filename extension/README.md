# TunaScript — VS Code Extension

Syntax highlighting, bracket matching, and snippets for [TunaScript](https://github.com/mayveiscous/tunascript) `.tuna` files.

---

## Features

- **Syntax highlighting** for keywords, types, operators, strings, numbers, booleans, and comments
- **Function highlighting** — `swim` declarations and call sites are coloured distinctly
- **Namespace highlighting** — `math.sqrt`, `string.upper`, `array.push` etc. highlight the namespace and method separately
- **Module keywords** — `from`, `cast` highlighted as import/export
- **Escape sequences** inside strings (`\n`, `\t`, `\"`, `\\`)
- **Bracket/paren/quote auto-closing**
- **Snippets** for all common patterns

---

## Snippets

| Prefix       | Description                          |
|--------------|--------------------------------------|
| `swim`       | Function definition                  |
| `swimr`      | Function with typed param and return |
| `if`         | If statement                         |
| `ife`        | If/else statement                    |
| `while`      | While loop                           |
| `for`        | For-in loop                          |
| `catch`      | Variable declaration                 |
| `catcht`     | Typed variable declaration           |
| `anchor`     | Constant declaration                 |
| `obj`        | Object literal                       |
| `from`       | Import from another `.tuna` file     |
| `castswim`   | Export a function                    |
| `castcatch`  | Export a variable                    |
| `bubble`     | Print to console                     |

---

## Language Quick Reference

```
><> this is a comment

><> variables
catch name: string = "tuna"
anchor MAX: number = 100

><> functions
swim add(a: number, b: number): number
  serve a + b
shore

><> control flow
if score >= 90
  bubble("A")
else
  bubble("F")
shore

while count < 10
  count++
shore

for fish in ["salmon", "tuna", "cod"]
  bubble(fish)
shore

><> objects
catch point: object = { x: 10, y: 20 }
bubble(point.x)

><> stdlib namespaces
bubble(math.sqrt(144))
bubble(string.upper("tuna"))
bubble(array.first([1, 2, 3]))

><> modules
from "utils.tuna" catch add, greet as hello
```

---

## Installation

Install from the [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=mayveiscous.tunascript) or download the `.vsix` and run:

```
code --install-extension tunascript-*.vsix
```