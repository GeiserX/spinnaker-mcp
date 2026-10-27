#!/usr/bin/env node
"use strict";

// Best effort at install time; run.js downloads on first run when install scripts are
// skipped (npm 12 default) or this fails on a flaky network. A checksum mismatch is a
// tampered or corrupt download, so that one fails the install.
require("./install.js")
  .ensureBinary({ log: (m) => console.log(m) })
  .catch((err) => {
    if (err.code === "EINTEGRITY") {
      console.error(`Failed to install spinnaker-mcp: ${err.message}`);
      process.exit(1);
    }
    console.error(`spinnaker-mcp: binary not downloaded yet (${err.message}); it will be fetched on first run.`);
  });
