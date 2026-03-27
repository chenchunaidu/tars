#!/usr/bin/env node
'use strict';

const cp = require('child_process');
const fs = require('fs');
const path = require('path');

const arch =
  process.arch === 'x64' ? 'amd64' : process.arch === 'arm64' ? 'arm64' : null;
if (!arch) {
  console.error('tars: unsupported architecture:', process.arch);
  process.exit(1);
}

const ext = process.platform === 'win32' ? '.exe' : '';
const bin = path.join(__dirname, '..', 'libexec', `tars-${arch}${ext}`);
if (!fs.existsSync(bin)) {
  console.error('tars: missing binary for', arch);
  process.exit(1);
}

const r = cp.spawnSync(bin, process.argv.slice(2), { stdio: 'inherit' });
if (r.error) {
  throw r.error;
}
process.exit(r.status === null ? 1 : r.status);
