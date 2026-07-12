#!/usr/bin/env node
'use strict';

const path = require('node:path');
const { spawnSync } = require('node:child_process');

const bin = path.join(__dirname, '..', process.platform === 'win32' ? 'runtui.exe' : 'runtui');

// stdio: 'inherit' hands the real TTY to the Go binary — required for the TUI.
const result = spawnSync(bin, process.argv.slice(2), { stdio: 'inherit' });

if (result.error) {
  console.error('runtui: binary not found — reinstall with: npm rebuild runtui');
  process.exit(1);
}
process.exit(result.status ?? 1);
