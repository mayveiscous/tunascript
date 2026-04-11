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
2. Find the executable and copy its path, something like the following:

```bash
   C:/Program Files/Tunascript/bin/
```

3. Open your Edit Environment Variables and navigate to User Variables
4. Edit your User Variables and create a new variable 
5. Paste the path to the executable into that variable and save the changes.

---

## Usage

Once you've installed tunascript and added it to your Path, you can run any .tuna file with the following command.

```bash
tunascript swim file.tuna
```

---

## Syntax Overview

### Variables

```tuna
catch name: string = "tuna"
```

### Functions

```tuna
swim add(a: number, b: number): number
  serve a + b
shore
```

### Conditionals

```tuna
catch isAwesome: bool = true
if isAwesome
  bubble("Tunascript is awesome!")
else
  ><> this will never happen!
shore
```

### Loops

```tuna
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
