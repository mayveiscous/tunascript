# TunaScript

TunaScript is a lightweight scripting language built to learn Go. It has no grand ambitions — but it turns out the ocean-themed syntax is surprisingly readable, even to people who don't write code.

TunaScript is licensed under the MIT License.

---

## Installation

1. Download the latest release
2. Copy the path to the executable, something like:

```
C:/Program Files/TunaScript/bin/
```

3. Open **Edit Environment Variables** and navigate to **User Variables**
4. Edit your `Path` variable and add the path above
5. Save and open a new terminal

---

## Usage

Run any `.tuna` file with:

```bash
tunascript run file.tuna
```

---

## Syntax Overview

### Variables & Constants

```tuna
catch name: string = "tuna"
catch age: number = 3
anchor MAX: number = 100    ><> anchor = constant, cannot be reassigned

><> type annotation is optional when a value is provided
catch inferred = "hello"
```

### Functions

```tuna
swim add(a: number, b: number): number
  serve a + b
shore

swim greet(who: string): string
  serve "hello, " + who + "!"
shore

bubble(add(3, 7))       ><> 10
bubble(greet("tuna"))   ><> hello, tuna!
```

### Conditionals

```tuna
catch score: number = 74

if score >= 90
  bubble("A")
else if score >= 70
  bubble("C")
else
  bubble("F")
shore
```

### Loops

```tuna
catch i: number = 0
while i < 5
  bubble(i)
  i++
shore

catch fish: []string = ["salmon", "mackerel", "tuna"]
for f in fish
  if f == "tuna"
    bubble("found the tuna!")
  shore
shore
```

### Objects

```tuna
catch point: object = { x: 10, y: 20 }
bubble(point.x)   ><> 10

point.x = 99
bubble(point.x)   ><> 99
```

### Standard Library

The stdlib is organised into three namespace objects and a few universal functions.

```tuna
><> math
bubble(math.sqrt(144))         ><> 12
bubble(math.pow(2, 8))         ><> 256
bubble(math.abs(-5))           ><> 5
bubble(math.min(3, 7))         ><> 3
bubble(math.randInt(1, 10))    ><> random number between 1 and 10
bubble(math.pi)                ><> 3.141592653589793

><> string
bubble(string.upper("tuna"))              ><> TUNA
bubble(string.replace("hi", "hi", "hey")) ><> hey
bubble(string.split("a,b,c", ","))        ><> [a, b, c]
bubble(string.startsWith("tuna", "tu"))   ><> true

><> array
bubble(array.first([1, 2, 3]))    ><> 1
bubble(array.last([1, 2, 3]))     ><> 3
bubble(array.sort([3, 1, 2]))     ><> [1, 2, 3]
bubble(array.reverse([1, 2, 3]))  ><> [3, 2, 1]
catch nums = array.push([1, 2], 3) ><> [1, 2, 3]

><> globals
bubble(len("tuna"))     ><> 4
bubble(typeOf(42))      ><> number
bubble(toString(3.14))  ><> 3.14
bubble(toNumber("99"))  ><> 99
```

### Modules

Split your code across multiple files using `cast` to export and `from`/`catch`/`as` to import.

```tuna
><> math_utils.tuna
cast swim square(n: number): number
  serve n * n
shore

cast anchor TAX_RATE: number = 0.08
```

```tuna
><> main.tuna
from "math_utils.tuna" catch square, TAX_RATE

bubble(square(5))  ><> 25
bubble(TAX_RATE)   ><> 0.08

><> import with an alias
from "math_utils.tuna" catch square as sq
bubble(sq(12))     ><> 144
```

---

## VS Code Extension

Syntax highlighting and snippets are available as a VS Code extension:
**[TunaScript — VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=mayveiscous.tunascript)**

---

## Issues

If you find a bug, open an issue and include:

- What happened
- Expected behavior
- Steps to reproduce