'use strict';

const assert = require('node:assert');
const { assetName } = require('./install.js');

assert.equal(assetName('darwin', 'arm64'), 'runtui_darwin_arm64.tar.gz');
assert.equal(assetName('darwin', 'x64'), 'runtui_darwin_amd64.tar.gz');
assert.equal(assetName('linux', 'arm64'), 'runtui_linux_arm64.tar.gz');
assert.equal(assetName('linux', 'x64'), 'runtui_linux_amd64.tar.gz');
assert.equal(assetName('win32', 'x64'), 'runtui_windows_amd64.zip');
assert.equal(assetName('win32', 'arm64'), null, 'windows arm64 not built');
assert.equal(assetName('freebsd', 'x64'), null, 'unsupported OS');

console.log('ok: asset mapping');
