#!/usr/bin/env node
"use strict";

const { execFileSync } = require("child_process");
const path = require("path");
const fs = require("fs");

const BIN_NAME = process.platform === "win32" ? "spinnaker-mcp.exe" : "spinnaker-mcp";
const BIN_PATH = path.join(__dirname, "bin", BIN_NAME);

if (!fs.existsSync(BIN_PATH)) {
  console.error("spinnaker-mcp binary not found. Run: npm run postinstall");
  process.exit(1);
}

const env = Object.assign({}, process.env, { TRANSPORT: "stdio" });

try {
  execFileSync(BIN_PATH, process.argv.slice(2), { env, stdio: "inherit" });
} catch (err) {
  process.exit(err.status || 1);
}
