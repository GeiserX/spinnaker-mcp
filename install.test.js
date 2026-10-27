"use strict";

const assert = require("node:assert");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const test = require("node:test");

const install = require("./install.js");

const VERSION = require("./package.json").version;
const H1 = "a".repeat(64);
const H2 = "b".repeat(64);

test("expectedHash takes the exact file name, not a substring", () => {
  const text = [
    `${H1}  spinnaker-mcp_${VERSION}_linux_amd64.tar.gz.sbom.json`,
    `${H2}  spinnaker-mcp_${VERSION}_linux_amd64.tar.gz`,
  ].join("\n");
  assert.strictEqual(install.expectedHash(text, `spinnaker-mcp_${VERSION}_linux_amd64.tar.gz`), H2);
  assert.throws(() => install.expectedHash(text, "other.tar.gz"), /No checksum/);
  assert.throws(() => install.expectedHash(`nothex  x.tar.gz`, "x.tar.gz"), /No checksum/);
});

test("a package without checksums.txt refuses to install", () => {
  assert.throws(() => install.readChecksums(path.join(os.tmpdir(), "does-not-exist.txt")), /missing from the package/);
});

test("downloads refuse plain HTTP before touching the network", async () => {
  const dest = path.join(fs.mkdtempSync(path.join(os.tmpdir(), "smcp-")), "x");
  await assert.rejects(install.download("http://github.com/x", dest), /non-HTTPS/);
  assert.strictEqual(fs.existsSync(dest), false);
});

test("asset names match the goreleaser archives", () => {
  assert.strictEqual(install.assetName("linux", "x64"), `spinnaker-mcp_${VERSION}_linux_amd64.tar.gz`);
  assert.strictEqual(install.assetName("win32", "arm64"), `spinnaker-mcp_${VERSION}_windows_arm64.zip`);
  assert.throws(() => install.assetName("aix", "ppc64"), /Unsupported/);
});
