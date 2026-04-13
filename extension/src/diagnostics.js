'use strict';

const vscode = require('vscode');

const VALID_TYPES = new Set([
  'number', 'string', 'bool', 'array', 'object', 'function', 'void', 'null',
]);

const NAMESPACE_METHODS = {
  math:   new Set(['floor','ceil','round','abs','min','max','pow','sqrt','rand','randInt','pi','e']),
  string: new Set(['upper','lower','trim','split','contains','replace','startsWith','endsWith','repeat']),
  array:  new Set(['push','pop','sort','reverse','first','last','slice','contains','join']),
};

const GLOBAL_BUILTINS = new Set(['bubble','typeOf','toNumber','toString','len']);
const BLOCK_OPENERS = /^\s*(cast\s+)?(swim|if|else(\s+if\b.*)?|while|for\b)\b/;
const SHORE_LINE    = /^\s*shore\s*$/;

function checkInvalidTypes(lines, diagnostics) {
  const re = /:\s*(\[\])?([a-zA-Z_][a-zA-Z0-9_]*)/g;

  for (let i = 0; i < lines.length; i++) {
    const line = stripComment(stripStrings(lines[i]));
    let match;
    while ((match = re.exec(line)) !== null) {
      const typeName = match[2];
      if (!VALID_TYPES.has(typeName)) {
        const col = match.index + match[0].length - typeName.length;
        diagnostics.push(makeDiagnostic(
          i, col, col + typeName.length,
          `Unknown type "${typeName}". Valid types: ${[...VALID_TYPES].join(', ')}`,
          vscode.DiagnosticSeverity.Warning,
        ));
      }
    }
  }
}

function checkUnknownNamespaceMethods(lines, diagnostics) {
  const re = /\b(math|string|array)\.([a-zA-Z_][a-zA-Z0-9_]*)/g;

  for (let i = 0; i < lines.length; i++) {
    const line = stripComment(stripStrings(lines[i]));
    let match;
    while ((match = re.exec(line)) !== null) {
      const ns     = match[1];
      const method = match[2];
      if (!NAMESPACE_METHODS[ns].has(method)) {
        const col = match.index + ns.length + 1;
        diagnostics.push(makeDiagnostic(
          i, col, col + method.length,
          `"${method}" is not a method of the "${ns}" namespace`,
          vscode.DiagnosticSeverity.Error,
        ));
      }
    }
  }
}

function checkBuiltinArgCounts(lines, diagnostics) {
  const EXPECTED_ARGS = {
    'math.floor':        1, 'math.ceil':     1, 'math.round':      1,
    'math.abs':          1, 'math.sqrt':     1, 'math.min':        2,
    'math.max':          2, 'math.pow':      2, 'math.randInt':    2,
    'string.upper':      1, 'string.lower':  1, 'string.trim':     1,
    'string.split':      2, 'string.contains': 2, 'string.replace': 3,
    'string.startsWith': 2, 'string.endsWith': 2, 'string.repeat':  2,
    'array.push':        2, 'array.pop':     1, 'array.sort':      1,
    'array.reverse':     1, 'array.first':   1, 'array.last':      1,
    'array.slice':       3, 'array.contains': 2, 'array.join':      2,
    'typeOf':            1, 'toNumber':       1, 'toString':        1,
    'len':               1,
  };

  const callRe = /\b(?:(math|string|array)\.)?([a-zA-Z_][a-zA-Z0-9_]*)\s*\(([^)]*)\)/g;

  for (let i = 0; i < lines.length; i++) {
    const line = stripComment(stripStrings(lines[i]));
    let match;
    while ((match = callRe.exec(line)) !== null) {
      const ns      = match[1];
      const fn      = match[2];
      const rawArgs = match[3].trim();
      const key     = ns ? `${ns}.${fn}` : fn;

      if (!(key in EXPECTED_ARGS)) continue;

      const argCount = rawArgs === '' ? 0 : countArgs(rawArgs);
      const expected = EXPECTED_ARGS[key];

      if (argCount !== expected) {
        const col = match.index;
        diagnostics.push(makeDiagnostic(
          i, col, col + match[0].length,
          `"${key}" expects ${expected} argument${expected !== 1 ? 's' : ''}, but got ${argCount}`,
          vscode.DiagnosticSeverity.Warning,
        ));
      }
    }
  }
}

function checkUnmatchedShore(lines, diagnostics) {
  const stack = []; // { indent, lineIndex }

  for (let i = 0; i < lines.length; i++) {
    const raw    = lines[i];
    const indent = raw.match(/^(\s*)/)[1].length;
    const clean  = stripComment(raw).trimEnd();

    if (BLOCK_OPENERS.test(clean)) {
      stack.push({ indent, lineIndex: i });
      continue;
    }

    if (SHORE_LINE.test(clean)) {
      if (stack.length === 0) {
        diagnostics.push(makeDiagnostic(
          i, indent, indent + 5,
          '"shore" with no matching block opener (swim / if / else / while / for)',
          vscode.DiagnosticSeverity.Warning,
        ));
      } else {
        stack.pop();
      }
    }
  }

  for (const { indent, lineIndex } of stack) {
    const openerText = lines[lineIndex].trim();
    const kw         = openerText.match(/^(\w+)/)?.[1] ?? 'block';
    diagnostics.push(makeDiagnostic(
      lineIndex, indent, indent + kw.length,
      `"${kw}" block is never closed with "shore"`,
      vscode.DiagnosticSeverity.Warning,
    ));
  }
}

function checkServeOutsideSwim(lines, diagnostics) {
  let swimDepth = 0;

  for (let i = 0; i < lines.length; i++) {
    const clean = stripComment(lines[i]);

    if (/^\s*(cast\s+)?swim\b/.test(clean))  swimDepth++;
    if (SHORE_LINE.test(clean) && swimDepth > 0) swimDepth--;

    const serveMatch = clean.match(/^(\s*)(serve)\b/);
    if (serveMatch && swimDepth === 0) {
      const col = serveMatch[1].length;
      diagnostics.push(makeDiagnostic(
        i, col, col + 5,
        '"serve" used outside of a "swim" function',
        vscode.DiagnosticSeverity.Warning,
      ));
    }
  }
}

function checkBreakContinueOutsideLoop(lines, diagnostics) {
  let loopDepth = 0;

  for (let i = 0; i < lines.length; i++) {
    const clean = stripComment(lines[i]);

    if (/^\s*(cast\s+)?(while|for)\b/.test(clean)) loopDepth++;
    if (SHORE_LINE.test(clean) && loopDepth > 0) loopDepth--;

    const bcMatch = clean.match(/^(\s*)(break|continue)\b/);
    if (bcMatch && loopDepth === 0) {
      const col = bcMatch[1].length;
      const kw  = bcMatch[2];
      diagnostics.push(makeDiagnostic(
        i, col, col + kw.length,
        `"${kw}" used outside of a loop`,
        vscode.DiagnosticSeverity.Warning,
      ));
    }
  }
}

function checkUninitializedAnchor(lines, diagnostics) {
  const re = /^\s*anchor\s+[a-zA-Z_][a-zA-Z0-9_]*(\s*:\s*[a-zA-Z_][a-zA-Z0-9_]*)?\s*$/;

  for (let i = 0; i < lines.length; i++) {
    const clean = stripComment(lines[i]);
    if (re.test(clean)) {
      const col = clean.match(/^(\s*)/)[1].length;
      diagnostics.push(makeDiagnostic(
        i, col, col + 6,
        '"anchor" constant must be assigned a value',
        vscode.DiagnosticSeverity.Error,
      ));
    }
  }
}

function checkInvalidCast(lines, diagnostics) {
  const re = /^\s*cast\s+(\w+)/;

  for (let i = 0; i < lines.length; i++) {
    const clean = stripComment(lines[i]);
    const match = re.exec(clean);
    if (!match) continue;

    const next = match[1];
    if (!['swim', 'catch', 'anchor'].includes(next)) {
      const col = clean.indexOf(next, clean.indexOf('cast'));
      diagnostics.push(makeDiagnostic(
        i, col, col + next.length,
        `"cast" must be followed by "swim", "catch", or "anchor" — got "${next}"`,
        vscode.DiagnosticSeverity.Error,
      ));
    }
  }
}

function makeDiagnostic(line, colStart, colEnd, message, severity) {
  const range = new vscode.Range(line, colStart, line, colEnd);
  const d = new vscode.Diagnostic(range, message, severity);
  d.source = 'TunaScript';
  return d;
}

function stripStrings(line) {
  return line.replace(/"(?:[^"\\]|\\.)*"/g, match => ' '.repeat(match.length));
}

function stripComment(line) {
  const idx = line.indexOf('><>');
  return idx === -1 ? line : line.slice(0, idx);
}

function countArgs(raw) {
  let depth = 0, count = 1;
  for (const ch of raw) {
    if (ch === '(' || ch === '[' || ch === '{') depth++;
    else if (ch === ')' || ch === ']' || ch === '}') depth--;
    else if (ch === ',' && depth === 0) count++;
  }
  return count;
}

function getDiagnostics(document) {
  const diagnostics = [];
  const lines = [];
  for (let i = 0; i < document.lineCount; i++) {
    lines.push(document.lineAt(i).text);
  }

  checkInvalidTypes(lines, diagnostics);
  checkUnknownNamespaceMethods(lines, diagnostics);
  checkBuiltinArgCounts(lines, diagnostics);
  checkUnmatchedShore(lines, diagnostics);
  checkServeOutsideSwim(lines, diagnostics);
  checkBreakContinueOutsideLoop(lines, diagnostics);
  checkUninitializedAnchor(lines, diagnostics);
  checkInvalidCast(lines, diagnostics);

  return diagnostics;
}

function registerDiagnostics(context) {
  const collection = vscode.languages.createDiagnosticCollection('tunascript');

  function refresh(document) {
    if (document.languageId !== 'tunascript') return;
    collection.set(document.uri, getDiagnostics(document));
  }

  context.subscriptions.push(
    collection,
    vscode.workspace.onDidOpenTextDocument(refresh),
    vscode.workspace.onDidChangeTextDocument(e => refresh(e.document)),
    vscode.workspace.onDidCloseTextDocument(doc => collection.delete(doc.uri)),
  );

  vscode.workspace.textDocuments.forEach(refresh);
}

module.exports = { registerDiagnostics };