You are analyzing the Tunascript codebase to design a complete plan for rewriting the language server to use the actual Tunascript AST and semantic information instead of regex/text-pattern matching.

## Goal

Produce a detailed implementation plan only. Do NOT modify code yet.

The objective is to redesign the LSP architecture so that completions, hover information, symbol lookup, go-to-definition, type resolution, diagnostics, and future language features are powered by Tunas  cript's lexer, parser, AST, and semantic analysis.

The current LSP relies heavily on regex/string scanning and is beginning to hit limitations around:

* User-defined types
* Member completion
* Type inference
* Scope tracking
* Hover accuracy
* Go-to-definition
* Future refactoring support

The rewrite should move the LSP toward the architecture used by mature language servers.

---

## Files To Read

Primary files:

* `/src/lexer/**`
* `/src/parser/**`
* `/ignore/tunascript-extension/**`

You may also inspect the rest of `/src/**` if additional context is needed for semantic analysis, type information, runtime structures, or compiler architecture.

You do NOT need to read any files outside those locations.

Avoid spending time analyzing unrelated project files.

---

## What I Want From You

After reviewing the codebase, produce:

### 1. Current Architecture Analysis

Describe:

* How the current LSP works
* Where regex/text scanning is being used
* Current completion strategy
* Current hover strategy
* Current type discovery strategy
* Current limitations

---

### 2. AST Inventory

Determine:

* What AST nodes already exist
* What information is already available in the AST
* What information is missing for LSP purposes
* Whether source locations/ranges are available
* Whether node ownership/scopes are represented

Show examples.

---

### 3. Semantic Layer Assessment

Determine whether TunaScript already has:

* Symbol tables
* Scope tracking
* Type representations
* Type inference
* Declaration tracking
* Function signatures
* Struct/class/type definitions
* Namespace information

If these exist, explain how the LSP can reuse them.

If they do not exist, explain exactly what must be added.

---

### 4. Proposed LSP Rewrite Architecture

Design a new architecture based on:

Source Text
↓
Lexer
↓
Parser
↓
AST
↓
Semantic Analysis
↓
LSP Features

Explain:

* Document lifecycle
* AST caching
* Incremental reparsing strategy
* Symbol cache strategy
* Type cache strategy
* Error recovery strategy

---

### 5. Completion System Design

Explain how completion should work for:

#### Variables

Example:

```tunascript
catch x as 5
x|
```

#### Members

```tuna
catch btn as imui.button(...)
btn.|
```

#### User Types

```tuna
school Person
    name: string
    age: number
shore

catch p: Person as {}

p.|
```

#### Namespaces

```tuna
imui.|
```

#### Function Parameters

```tuna
swim test(player: Person)
    player.|
shore
```

Describe the exact resolution flow.

---

### 6. Hover System Design

Explain how hover should be derived from:

* AST nodes
* Symbol information
* Type information

Show example outputs.

---

### 7. Go-To-Definition Design

Explain:

* Declaration indexing
* Symbol lookup
* Scope resolution
* Cross-file support

---

### 8. Diagnostics Design

Explain how diagnostics should eventually come from:

* Parser errors
* Semantic errors
* Type errors

instead of regex validation.

---

### 9. Migration Strategy

Produce a phased migration plan.

The plan should minimize risk and avoid rewriting everything at once.

For example:

Phase 1:

* AST loading
* Symbol collection

Phase 2:

* Hover rewrite

Phase 3:

* Completion rewrite

etc.

Recommend the safest order.

---

### 10. Concrete File-Level Refactor Plan

List:

* Existing files that should be modified
* New files that should be added
* Responsibilities of each file

Provide a proposed directory structure.

---

## Important Constraints

* Do not propose adding external dependencies unless absolutely necessary.
* If an external dependency is deemed neccessary, stop and let me know, and we'll think of something else.
* Assume the current project philosophy favors **NO** dependencies.
* Prefer reusing existing Tunascript lexer/parser code wherever possible, duplicate code should be minimal.
* Prefer a clean architecture that will scale to future features such as:

  * Rename Symbol
  * Find References
  * Signature Help
  * Semantic Tokens
  * Refactoring Tools

Focus on producing the highest quality architecture plan possible before any implementation begins.