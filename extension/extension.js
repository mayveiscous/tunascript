'use strict';

const vscode = require('vscode');
const { registerHovers } = require('./src/hovers');
const { registerDiagnostics } = require('./src/diagnostics');
const { registerShoreInsertion } = require('./src/insertion');
const { registerCompletions } = require('./src/completions');

function activate(context) {
  registerHovers(context);
  registerDiagnostics(context);
  registerShoreInsertion(context);
  registerCompletions(context);
}

function deactivate() {}

module.exports = { activate, deactivate };