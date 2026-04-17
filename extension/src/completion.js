'use strict';

const vscode = require('vscode');

const NAMESPACE_COMPLETIONS = {
  math: [
    { label: 'floor',   snippet: 'floor(${1:n})',           doc: 'Rounds down to the nearest integer.' },
    { label: 'ceil',    snippet: 'ceil(${1:n})',            doc: 'Rounds up to the nearest integer.' },
    { label: 'round',   snippet: 'round(${1:n})',           doc: 'Rounds to the nearest integer.' },
    { label: 'abs',     snippet: 'abs(${1:n})',             doc: 'Returns the absolute value.' },
    { label: 'min',     snippet: 'min(${1:a}, ${2:b})',     doc: 'Returns the smaller of two numbers.' },
    { label: 'max',     snippet: 'max(${1:a}, ${2:b})',     doc: 'Returns the larger of two numbers.' },
    { label: 'pow',     snippet: 'pow(${1:base}, ${2:exp})',doc: 'Returns base raised to the power of exp.' },
    { label: 'sqrt',    snippet: 'sqrt(${1:n})',            doc: 'Returns the square root.' },
    { label: 'rand',    snippet: 'rand()',                  doc: 'Returns a random float in [0, 1).' },
    { label: 'randInt', snippet: 'randInt(${1:min}, ${2:max})', doc: 'Returns a random integer in [min, max] inclusive.' },
    { label: 'pi',      snippet: 'pi',                     doc: 'The mathematical constant π ≈ 3.14159.' },
    { label: 'e',       snippet: 'e',                      doc: "Euler's number e ≈ 2.71828." },
  ],
  string: [
    { label: 'upper',      snippet: 'upper(${1:s})',                        doc: 'Converts the string to uppercase.' },
    { label: 'lower',      snippet: 'lower(${1:s})',                        doc: 'Converts the string to lowercase.' },
    { label: 'trim',       snippet: 'trim(${1:s})',                         doc: 'Removes leading and trailing whitespace.' },
    { label: 'split',      snippet: 'split(${1:s}, "${2:sep}")',            doc: 'Splits a string by separator into an array.' },
    { label: 'contains',   snippet: 'contains(${1:s}, "${2:sub}")',         doc: 'Returns true if s contains sub.' },
    { label: 'replace',    snippet: 'replace(${1:s}, "${2:old}", "${3:new}")', doc: 'Replaces all occurrences of old with new.' },
    { label: 'startsWith', snippet: 'startsWith(${1:s}, "${2:prefix}")',    doc: 'Returns true if s starts with prefix.' },
    { label: 'endsWith',   snippet: 'endsWith(${1:s}, "${2:suffix}")',      doc: 'Returns true if s ends with suffix.' },
    { label: 'repeat',     snippet: 'repeat(${1:s}, ${2:n})',               doc: 'Repeats s exactly n times.' },
  ],
  array: [
    { label: 'push',     snippet: 'push(${1:arr}, ${2:value})',        doc: '⚠️ Mutating. Appends value to the end of arr in place.' },
    { label: 'pop',      snippet: 'pop(${1:arr})',                     doc: '⚠️ Mutating. Removes the last element of arr in place.' },
    { label: 'sort',     snippet: 'sort(${1:arr})',                    doc: '⚠️ Mutating. Sorts arr in place.' },
    { label: 'reverse',  snippet: 'reverse(${1:arr})',                 doc: '⚠️ Mutating. Reverses arr in place.' },
    { label: 'first',    snippet: 'first(${1:arr})',                   doc: '⚠️ Mutating. Truncates arr to its first element.' },
    { label: 'last',     snippet: 'last(${1:arr})',                    doc: '⚠️ Mutating. Truncates arr to its last element.' },
    { label: 'slice',    snippet: 'slice(${1:arr}, ${2:start}, ${3:end})', doc: '⚠️ Mutating. Truncates arr to range [start, end).' },
    { label: 'contains', snippet: 'contains(${1:arr}, ${2:value})',   doc: 'Returns true if the array contains value.' },
    { label: 'join',     snippet: 'join(${1:arr}, "${2:sep}")',        doc: 'Joins all elements into a string separated by sep.' },
  ],
};

const GLOBAL_BUILTINS = [
  { label: 'bubble',   snippet: 'bubble($0)',              doc: 'Prints all arguments to stdout, space-separated.' },
  { label: 'typeOf',   snippet: 'typeOf(${1:value})',      doc: 'Returns the runtime type of a value as a string.' },
  { label: 'toNumber', snippet: 'toNumber(${1:value})',    doc: 'Converts a string, bool, or number to a number.' },
  { label: 'toString', snippet: 'toString(${1:value})',    doc: 'Converts any value to its string representation.' },
  { label: 'len',      snippet: 'len(${1:value})',         doc: 'Returns the length of a string or array.' },
];

const KEYWORD_COMPLETIONS = [
  { label: 'swim',     snippet: 'swim ${1:name}(${2:params}): ${3:void}\n\t$0\nshore', doc: 'Define a function.' },
  { label: 'catch',    snippet: 'catch ${1:name} = ${2:value}',                        doc: 'Declare a mutable variable.' },
  { label: 'anchor',   snippet: 'anchor ${1:NAME} = ${2:value}',                      doc: 'Declare an immutable constant.' },
  { label: 'serve',    snippet: 'serve $0',                                            doc: 'Return a value from a swim function.' },
  { label: 'shore',    snippet: 'shore',                                               doc: 'Close a block (swim, if, while, for).' },
  { label: 'if',       snippet: 'if ${1:condition}\n\t$0\nshore',                     doc: 'Conditional block.' },
  { label: 'else',     snippet: 'else\n\t$0',                                          doc: 'Fallback branch of an if block.' },
  { label: 'while',    snippet: 'while ${1:condition}\n\t$0\nshore',                  doc: 'Loop while condition is truthy.' },
  { label: 'for',      snippet: 'for ${1:item} in ${2:collection}\n\t$0\nshore',      doc: 'Iterate over every element in an array.' },
  { label: 'from',     snippet: 'from "${1:module.tuna}" catch ${2:name}',            doc: 'Import from another .tuna file.' },
  { label: 'cast',     snippet: 'cast ',                                               doc: 'Export a variable or function.' },
  { label: 'break',    snippet: 'break',                                               doc: 'Exit the current loop.' },
  { label: 'continue', snippet: 'continue',                                            doc: 'Skip to the next loop iteration.' },
  { label: 'nil',      snippet: 'nil',                                                 doc: 'The null/empty value.' },
  { label: 'true',     snippet: 'true',                                                doc: 'Boolean literal true.' },
  { label: 'false',    snippet: 'false',                                               doc: 'Boolean literal false.' },
  { label: 'typeof',   snippet: 'typeof ${1:value}',                                  doc: 'Keyword type inspection.' },
  { label: 'and',      snippet: 'and',                                                 doc: 'Logical AND operator.' },
  { label: 'or',       snippet: 'or',                                                  doc: 'Logical OR operator.' },
  { label: 'new',      snippet: 'new ${1:ClassName}($0)',                              doc: 'Create a new class instance.' },
  { label: 'as',       snippet: 'as ${1:alias}',                                       doc: 'Rename an import.' },
  { label: 'in',       snippet: 'in',                                                  doc: 'Used in for loops.' },
];

const NAMESPACE_TRIGGERS = [
  { label: 'math',   doc: 'Built-in math namespace. Type `math.` to see methods.' },
  { label: 'string', doc: 'Built-in string namespace. Type `string.` to see methods.' },
  { label: 'array',  doc: 'Built-in array namespace. Type `array.` to see methods.' },
];

function makeItem(label, snippet, doc, kind) {
  const item = new vscode.CompletionItem(label, kind);
  item.insertText = new vscode.SnippetString(snippet);
  item.documentation = new vscode.MarkdownString(doc);
  // Ensure it shows up even without the usual word-boundary trigger
  item.preselect = false;
  return item;
}

function registerCompletions(context) {
  const nsProvider = vscode.languages.registerCompletionItemProvider(
    'tunascript',
    {
      provideCompletionItems(document, position) {
        const linePrefix = document.lineAt(position).text.slice(0, position.character);
        const nsMatch = linePrefix.match(/(math|string|array)\.$/);
        if (!nsMatch) return undefined;

        const ns = nsMatch[1];
        return NAMESPACE_COMPLETIONS[ns].map(({ label, snippet, doc }) =>
          makeItem(label, snippet, doc, vscode.CompletionItemKind.Method)
        );
      },
    },
    '.'
  );

  const globalProvider = vscode.languages.registerCompletionItemProvider(
    'tunascript',
    {
      provideCompletionItems(document, position) {
        const linePrefix = document.lineAt(position).text.slice(0, position.character);
        // Don't fire inside a namespace call (e.g. after "math.")
        if (/(math|string|array)\.$/.test(linePrefix)) return undefined;

        const items = [];

        for (const { label, snippet, doc } of GLOBAL_BUILTINS) {
          items.push(makeItem(label, snippet, doc, vscode.CompletionItemKind.Function));
        }

        for (const { label, snippet, doc } of KEYWORD_COMPLETIONS) {
          items.push(makeItem(label, snippet, doc, vscode.CompletionItemKind.Keyword));
        }

        for (const { label, doc } of NAMESPACE_TRIGGERS) {
          const item = new vscode.CompletionItem(label, vscode.CompletionItemKind.Module);
          item.documentation = new vscode.MarkdownString(doc);
          item.commitCharacters = ['.'];
          items.push(item);
        }

        return items;
      },
    }
  );

  context.subscriptions.push(nsProvider, globalProvider);
}

module.exports = { registerCompletions };