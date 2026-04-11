# Tunascript

Tunascript. Truly a scripting language of all time. Developed solely for the purpose of learning Golang, it servers no immediate purpose.
Tunascript is licensed under the MIT License.

---

## Features

* Fun syntax
* Lightweight and portable
  
---

## Installation

1. Download the latest release
3. Add the file to your PATH

---

## Usage

Once you've installed tunascript and added it to your Path, you can run any .tuna file with the following command.

```bash
tunascript swim file.tuna
```

---

## Syntax Overview

### Variables

```ts
catch name: string = "tuna"
```

### Functions

```ts
swim add(a: number, b: number): number
  serve a + b
shore
```

### Conditionals

```ts
catch isAwesome: bool = true
if isAwesome
  bubble("Tunascript is awesome!")
else
  ><> this will never happen!
shore
```

### Loops

```ts
catch i: number = 0
while i < 10
  bubble(i)
shore

catch fish: []string = ["salmon", "mackeral", "tuna"]
for f in fish
  if f == "tuna"
    bubble("it's a tuna!")
  shore
shore
```

---

## Issues

If you find a bug, open an issue and include:

* What happened
* Expected behavior
* Steps to reproduce
