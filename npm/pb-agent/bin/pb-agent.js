#!/usr/bin/env node

"use strict";

const { spawnSync } = require("node:child_process");

const packageName = `pb-agent-${process.platform}-${process.arch}`;
let binary;

try {
  binary = require.resolve(`${packageName}/bin/pb-agent${process.platform === "win32" ? ".exe" : ""}`);
} catch {
  process.stderr.write(
    `pb-agent: native package ${packageName} is unavailable for this platform.\n`,
  );
  process.exit(1);
}

const result = spawnSync(binary, process.argv.slice(2), { stdio: "inherit" });
if (result.error) {
  process.stderr.write(`pb-agent: ${result.error.message}\n`);
  process.exit(1);
}
process.exit(result.status ?? 1);
