'use strict';

// Postinstall: download the prebuilt runtui binary for this platform
// from GitHub Releases and extract it into the package directory.

const fs = require('node:fs');
const path = require('node:path');
const { spawnSync } = require('node:child_process');

const REPO = 'pedromendonka/runtui';

// Keys are `${process.platform}-${process.arch}`; values are GoReleaser
// archive names (see .goreleaser.yml name_template). Node "x64" = Go "amd64".
const ASSETS = {
  'darwin-arm64': 'runtui_darwin_arm64.tar.gz',
  'darwin-x64': 'runtui_darwin_amd64.tar.gz',
  'linux-arm64': 'runtui_linux_arm64.tar.gz',
  'linux-x64': 'runtui_linux_amd64.tar.gz',
  'win32-x64': 'runtui_windows_amd64.zip',
};

function assetName(platform, arch) {
  return ASSETS[`${platform}-${arch}`] || null;
}

function fail(msg) {
  console.error(`runtui: ${msg}`);
  process.exit(1);
}

async function main() {
  const { version } = require('./package.json');
  if (version === '0.0.0') {
    console.log('runtui: dev version (0.0.0), skipping binary download');
    return;
  }

  const asset = assetName(process.platform, process.arch);
  if (!asset) {
    fail(
      `unsupported platform ${process.platform}-${process.arch}\n` +
        'Try: go install github.com/pedromendonka/runtui@latest'
    );
  }

  const url = `https://github.com/${REPO}/releases/download/v${version}/${asset}`;
  const res = await fetch(url);
  if (!res.ok) {
    fail(`download failed (HTTP ${res.status}): ${url}`);
  }

  const archive = path.join(__dirname, asset);
  fs.writeFileSync(archive, Buffer.from(await res.arrayBuffer()));

  // ponytail: spawn tar instead of a JS extract dep — present on macOS,
  // Linux, and Windows 10+ (bsdtar, which also reads zip).
  const tar = spawnSync('tar', ['-xf', archive], { cwd: __dirname, stdio: 'inherit' });
  fs.unlinkSync(archive);
  if (tar.error || tar.status !== 0) {
    fail('failed to extract binary (is `tar` on PATH?)');
  }

  const bin = path.join(__dirname, process.platform === 'win32' ? 'runtui.exe' : 'runtui');
  if (!fs.existsSync(bin)) {
    fail(`archive did not contain expected binary: ${bin}`);
  }
  if (process.platform !== 'win32') {
    fs.chmodSync(bin, 0o755);
  }
  console.log(`runtui ${version} installed`);
}

module.exports = { assetName };

if (require.main === module) {
  main().catch((err) => fail(err.message));
}
