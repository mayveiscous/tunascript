'use strict';

const vscode = require('vscode');

// ── Hover documentation map ───────────────────────────────────────────────────

const HOVER_DOCS = {
  // ── Global functions ──────────────────────────────────────────────────────
  bubble:
    '**`bubble(...values)`**\n\nPrints all arguments to stdout, space-separated.\n\n```tuna\nbubble("hello", name)\n```',

  typeOf:
    '**`typeOf(value): string`**\n\nReturns the runtime type of a value as a string.\n\nPossible values: `"number"` `"string"` `"bool"` `"array"` `"object"` `"function"` `"null"`\n\n```tuna\ntypeOf(42)      ><> "number"\ntypeOf("hi")   ><> "string"\n```',

  toNumber:
    '**`toNumber(value): number`**\n\nConverts a `string`, `bool`, or `number` to a number.\n\n- `true → 1`, `false → 0`\n- Panics if the string is not numeric.\n\n```tuna\ntoNumber("3.14")  ><> 3.14\ntoNumber(true)    ><> 1\n```',

  toString:
    '**`toString(value): string`**\n\nConverts any value to its string representation.\n\n```tuna\ntoString(100)    ><> "100"\ntoString(true)   ><> "true"\n```',

  len:
    '**`len(value): number`**\n\nReturns the length of a `string` (Unicode-aware) or `array`.\n\n```tuna\nlen("tuna")        ><> 4\nlen([1, 2, 3])     ><> 3\n```',

  // ── Keywords ──────────────────────────────────────────────────────────────
  catch:
    '**`catch name = value`**\n\nDeclares a mutable variable. Optionally typed:\n\n```tuna\ncatch name = "tuna"\ncatch score: number = 0\n```',

  anchor:
    '**`anchor NAME = value`**\n\nDeclares an immutable constant. Must be assigned a value.\n\n```tuna\nanchor MAX: number = 100\n```',

  swim:
    '**`swim name(params): returnType`**\n\nDefines a function. The body is closed with `shore`.\n\n```tuna\nswim add(a: number, b: number): number\n  serve a + b\nshore\n```',

  serve:
    '**`serve expr`**\n\nReturns a value from the current `swim` function.\n\n```tuna\nswim double(n: number): number\n  serve n * 2\nshore\n```',

  shore:
    '**`shore`**\n\nCloses a block. Used to end `swim`, `if`, `else`, `while`, and `for` blocks.\n\n```tuna\nif x > 0\n  bubble("positive")\nshore\n```',

  if:
    '**`if condition`**\n\nConditional block, closed with `shore`. Can be followed by `else` or `else if`.\n\n```tuna\nif score >= 90\n  bubble("A")\nelse\n  bubble("F")\nshore\n```',

  else:
    '**`else`**\n\nFallback branch of an `if` block. Can also be `else if condition`.\n\n```tuna\nif x > 0\n  bubble("pos")\nelse\n  bubble("neg")\nshore\n```',

  while:
    '**`while condition`**\n\nLoops while the condition is truthy. Closed with `shore`.\n\n```tuna\nwhile count < 10\n  count++\nshore\n```',

  for:
    '**`for item in collection`**\n\nIterates over every element in an array. Closed with `shore`.\n\n```tuna\nfor fish in ["salmon", "tuna", "cod"]\n  bubble(fish)\nshore\n```',

  in:
    '**`in`**\n\nUsed in `for` loops to specify the collection being iterated.\n\n```tuna\nfor item in myArray\n```',

  break:
    '**`break`**\n\nImmediately exits the current `while` or `for` loop.\n\n```tuna\nwhile true\n  if done\n    break\n  shore\nshore\n```',

  continue:
    '**`continue`**\n\nSkips the rest of the current loop iteration and moves to the next.\n\n```tuna\nfor n in nums\n  if n == 0\n    continue\n  shore\n  bubble(n)\nshore\n```',

  from:
    '**`from "file.tuna" catch name`**\n\nImports exported values from another `.tuna` file. Supports aliasing with `as`.\n\n```tuna\nfrom "utils.tuna" catch add, greet as hello\n```',

  cast:
    '**`cast`**\n\nExports a variable or function from this module.\n\n```tuna\ncast catch PI = 3.14\ncast swim greet(name: string): void\n  bubble("hi", name)\nshore\n```',

  as:
    '**`as alias`**\n\nRenames an import when using `from`.\n\n```tuna\nfrom "utils.tuna" catch greet as hello\n```',

  new:
    '**`new ClassName(args)`**\n\nCreates a new class instance.\n\n```tuna\ncatch p = new Point(10, 20)\n```',

  typeof:
    '**`typeof value`**\n\nKeyword form of type inspection. Returns type as a string. Same as `typeOf()`.\n\n```tuna\ntypeof x   ><> "number"\n```',

  and:
    '**`and`**\n\nLogical AND operator.\n\n```tuna\nif x > 0 and x < 10\n```',

  or:
    '**`or`**\n\nLogical OR operator.\n\n```tuna\nif x == 0 or x == nil\n```',

  nil:
    '**`nil`**\n\nThe null/empty value. Use `== nil` to check for absence.\n\n```tuna\nif result == nil\n  bubble("nothing")\nshore\n```',

  true:  '**`true`**\n\nBoolean literal `true`.',
  false: '**`false`**\n\nBoolean literal `false`.',

  // ── math namespace ────────────────────────────────────────────────────────
  'math.floor':
    '**`math.floor(n: number): number`**\n\nRounds down to the nearest integer.\n\n```tuna\nmath.floor(4.9)  ><> 4\n```',

  'math.ceil':
    '**`math.ceil(n: number): number`**\n\nRounds up to the nearest integer.\n\n```tuna\nmath.ceil(4.1)  ><> 5\n```',

  'math.round':
    '**`math.round(n: number): number`**\n\nRounds to the nearest integer.\n\n```tuna\nmath.round(4.5)  ><> 5\n```',

  'math.abs':
    '**`math.abs(n: number): number`**\n\nReturns the absolute value.\n\n```tuna\nmath.abs(-7)  ><> 7\n```',

  'math.min':
    '**`math.min(a: number, b: number): number`**\n\nReturns the smaller of two numbers.\n\n```tuna\nmath.min(3, 7)  ><> 3\n```',

  'math.max':
    '**`math.max(a: number, b: number): number`**\n\nReturns the larger of two numbers.\n\n```tuna\nmath.max(3, 7)  ><> 7\n```',

  'math.pow':
    '**`math.pow(base: number, exp: number): number`**\n\nReturns `base` raised to the power of `exp`.\n\n```tuna\nmath.pow(2, 8)  ><> 256\n```',

  'math.sqrt':
    '**`math.sqrt(n: number): number`**\n\nReturns the square root.\n\n```tuna\nmath.sqrt(144)  ><> 12\n```',

  'math.rand':
    '**`math.rand(): number`**\n\nReturns a random float in `[0, 1)`.\n\n```tuna\ncatch roll = math.rand()\n```',

  'math.randInt':
    '**`math.randInt(min: number, max: number): number`**\n\nReturns a random integer in `[min, max]` **inclusive**.\n\n```tuna\nmath.randInt(1, 6)  ><> dice roll\n```',

  'math.pi':
    '**`math.pi`**\n\nThe mathematical constant π ≈ 3.14159265358979.',

  'math.e':
    '**`math.e`**\n\nEuler\'s number e ≈ 2.71828182845905.',

  // ── string namespace ──────────────────────────────────────────────────────
  'string.upper':
    '**`string.upper(s: string): string`**\n\nConverts the string to uppercase.\n\n```tuna\nstring.upper("tuna")  ><> "TUNA"\n```',

  'string.lower':
    '**`string.lower(s: string): string`**\n\nConverts the string to lowercase.\n\n```tuna\nstring.lower("TUNA")  ><> "tuna"\n```',

  'string.trim':
    '**`string.trim(s: string): string`**\n\nRemoves leading and trailing whitespace.\n\n```tuna\nstring.trim("  hi  ")  ><> "hi"\n```',

  'string.split':
    '**`string.split(s: string, sep: string): []string`**\n\nSplits a string by separator into an array of strings.\n\n```tuna\nstring.split("a,b,c", ",")  ><> ["a","b","c"]\n```',

  'string.contains':
    '**`string.contains(s: string, sub: string): bool`**\n\nReturns `true` if `s` contains `sub`.\n\n```tuna\nstring.contains("tuna", "un")  ><> true\n```',

  'string.replace':
    '**`string.replace(s: string, old: string, new: string): string`**\n\nReplaces **all** occurrences of `old` with `new`.\n\n```tuna\nstring.replace("aabbcc", "bb", "XX")  ><> "aaXXcc"\n```',

  'string.startsWith':
    '**`string.startsWith(s: string, prefix: string): bool`**\n\nReturns `true` if `s` starts with `prefix`.\n\n```tuna\nstring.startsWith("tuna", "tu")  ><> true\n```',

  'string.endsWith':
    '**`string.endsWith(s: string, suffix: string): bool`**\n\nReturns `true` if `s` ends with `suffix`.\n\n```tuna\nstring.endsWith("tuna", "na")  ><> true\n```',

  'string.repeat':
    '**`string.repeat(s: string, n: number): string`**\n\nRepeats `s` exactly `n` times. `n` must be >= 0.\n\n```tuna\nstring.repeat("ab", 3)  ><> "ababab"\n```',

  // ── array namespace ───────────────────────────────────────────────────────
  'array.push':
    '**`array.push(arr, value)`** ⚠️ *Mutating*\n\nAppends `value` to the end of `arr` **in place**.\n\n```tuna\narray.push(myList, 42)\n```',

  'array.pop':
    '**`array.pop(arr)`** ⚠️ *Mutating*\n\nRemoves the last element of `arr` **in place**. Panics on an empty array.\n\n```tuna\narray.pop(myList)\n```',

  'array.sort':
    '**`array.sort(arr)`** ⚠️ *Mutating*\n\nSorts `arr` **in place**. Requires all numbers or all strings.\n\n```tuna\narray.sort(scores)\n```',

  'array.reverse':
    '**`array.reverse(arr)`** ⚠️ *Mutating*\n\nReverses `arr` **in place**.\n\n```tuna\narray.reverse(myList)\n```',

  'array.first':
    '**`array.first(arr)`** ⚠️ *Mutating*\n\nTruncates `arr` **in place** to only its first element. Panics on an empty array.',

  'array.last':
    '**`array.last(arr)`** ⚠️ *Mutating*\n\nTruncates `arr` **in place** to only its last element. Panics on an empty array.',

  'array.slice':
    '**`array.slice(arr, start: number, end: number)`** ⚠️ *Mutating*\n\nTruncates `arr` **in place** to the range `[start, end)`. Panics if out of bounds.\n\n```tuna\narray.slice(myList, 1, 3)\n```',

  'array.contains':
    '**`array.contains(arr, value): bool`**\n\nReturns `true` if the array contains `value`. Uses value equality.\n\n```tuna\narray.contains([1,2,3], 2)  ><> true\n```',

  'array.join':
    '**`array.join(arr, sep: string): string`**\n\nJoins all elements into a single string separated by `sep`.\n\n```tuna\narray.join(["a","b","c"], "-")  ><> "a-b-c"\n```',
};

// ── Namespace method resolver ─────────────────────────────────────────────────

function resolveHoverKey(document, position) {
  const wordRange = document.getWordRangeAtPosition(position);
  if (!wordRange) return null;

  const word = document.getText(wordRange);
  const lineText = document.lineAt(position.line).text;
  const before = lineText.slice(0, wordRange.start.character);
  const nsMatch = before.match(/(math|string|array)\.$/);

  return nsMatch ? `${nsMatch[1]}.${word}` : word;
}

// ── Registration ──────────────────────────────────────────────────────────────

function registerHovers(context) {
  context.subscriptions.push(
    vscode.languages.registerHoverProvider('tunascript', {
      provideHover(document, position) {
        const key = resolveHoverKey(document, position);
        if (!key || !HOVER_DOCS[key]) return null;

        const md = new vscode.MarkdownString(HOVER_DOCS[key]);
        md.isTrusted = true;
        return new vscode.Hover(md);
      },
    })
  );
}

module.exports = { registerHovers };