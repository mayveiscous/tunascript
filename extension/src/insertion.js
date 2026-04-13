'use strict';

const vscode = require('vscode');

const BLOCK_OPENER_RE = /^\s*(cast\s+)?(swim|if(\s+.*)?|else(\s+if\s+.*)?|while(\s+.*)?|for\s+\w+\s+in\s+.*)$/;

function registerShoreInsertion(context) {
  context.subscriptions.push(
    vscode.workspace.onDidChangeTextDocument(event => {
      const editor = vscode.window.activeTextEditor;
      if (!editor || event.document !== editor.document) return;
      if (editor.document.languageId !== 'tunascript') return;

      const change = event.contentChanges[0];
      if (!change || change.text !== '\n') return;

      const lineNum = change.range.start.line;
      const prevLine = editor.document.lineAt(lineNum).text;
      const indent   = prevLine.match(/^(\s*)/)[1];

      if (!BLOCK_OPENER_RE.test(prevLine)) return;

      editor.edit(edit => {
        const cursorPos = editor.selection.active;
        edit.insert(cursorPos, `\n${indent}shore`);
      }).then(() => {
        const bodyLine = editor.selection.active.line - 1;
        const bodyIndent = indent.length + 2;
        const newPos = new vscode.Position(bodyLine, bodyIndent);
        editor.selection = new vscode.Selection(newPos, newPos);
      });
    })
  );
}

module.exports = { registerShoreInsertion };