'use strict';

const vscode = require('vscode');
const { registerHovers } = require('./src/hovers');
const { registerDiagnostics } = require('./src/diagnostics');
const { registerShoreInsertion } = require('./src/insertion');

function activate(context) {
  registerHovers(context);
  registerDiagnostics(context);
  registerShoreInsertion(context);
}

function deactivate() {}

module.exports = { activate, deactivate };