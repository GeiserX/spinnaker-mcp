#!/usr/bin/env node
"use strict";

const { execFileSync } = require("child_process");
const { ensureBinary } = require("./install.js");

// stdout carries MCP messages, so download progress goes to stderr.
ensureBinary({ log: (m) => process.stderr.write(m + "\n") })
  .then((bin) => {
    const env = Object.assign({}, process.env, { TRANSPORT: "stdio" });
    try {
      execFileSync(bin, process.argv.slice(2), { env, stdio: "inherit" });
    } catch (err) {
      process.exit(err.status || 1);
    }
  })
  .catch((err) => {
    process.stderr.write(`spinnaker-mcp: ${err.message}\n`);
    process.exit(1);
  });
