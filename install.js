"use strict";

// Downloads the spinnaker-mcp release binary for this platform and checks it against the
// sha256 in checksums.txt, which the release workflow puts inside the npm package. The
// expected hash therefore comes from the package npm verified, not from the GitHub
// release the binary is fetched from, so swapping release assets cannot pass the check.
//
// postinstall.js calls this at install time; run.js calls it on first run, because npm 12,
// pnpm and Bun skip install scripts by default.

const { execFileSync } = require("child_process");
const crypto = require("crypto");
const fs = require("fs");
const https = require("https");
const path = require("path");

const VERSION = require("./package.json").version;
const REPO = "GeiserX/spinnaker-mcp";
const BIN_NAME = process.platform === "win32" ? "spinnaker-mcp.exe" : "spinnaker-mcp";
const BIN_DIR = path.join(__dirname, "bin");
const BIN_PATH = path.join(BIN_DIR, BIN_NAME);
const CHECKSUMS = path.join(__dirname, "checksums.txt");
const MAX_DOWNLOAD = 100 * 1024 * 1024; // the archives are about 6 MB
const TIMEOUT_MS = 60 * 1000;

function assetName(platform = process.platform, arch = process.arch) {
  const p = { darwin: "darwin", linux: "linux", win32: "windows" }[platform];
  const a = { x64: "amd64", arm64: "arm64" }[arch];
  if (!p || !a) {
    throw new Error(`Unsupported platform: ${platform}-${arch}`);
  }
  return `spinnaker-mcp_${VERSION}_${p}_${a}.${p === "windows" ? "zip" : "tar.gz"}`;
}

// expectedHash returns the sha256 listed for exactly this file name.
function expectedHash(checksumsText, name) {
  for (const line of checksumsText.split("\n")) {
    const [hash, file] = line.trim().split(/\s+/);
    if (file === name && /^[0-9a-f]{64}$/.test(hash)) {
      return hash;
    }
  }
  throw new Error(`No checksum for ${name} in the bundled checksums.txt`);
}

function readChecksums(file = CHECKSUMS) {
  try {
    return fs.readFileSync(file, "utf8");
  } catch {
    throw new Error("checksums.txt is missing from the package; refusing to install an unverified binary");
  }
}

// download streams url to dest over HTTPS only (redirects included), with a size cap,
// and resolves to the sha256 of what it wrote.
function download(url, dest, redirects = 5) {
  return new Promise((resolve, reject) => {
    let parsed;
    try {
      parsed = new URL(url);
    } catch (e) {
      return reject(e);
    }
    if (parsed.protocol !== "https:") {
      return reject(new Error(`Refusing non-HTTPS download URL: ${url}`));
    }
    const req = https.get(parsed, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        res.resume();
        if (redirects <= 0) {
          return reject(new Error("Too many redirects"));
        }
        return download(new URL(res.headers.location, parsed).toString(), dest, redirects - 1).then(resolve, reject);
      }
      if (res.statusCode !== 200) {
        res.resume();
        return reject(new Error(`Download failed: HTTP ${res.statusCode}`));
      }
      const hash = crypto.createHash("sha256");
      const out = fs.createWriteStream(dest, { mode: 0o600 });
      let size = 0;
      res.on("data", (chunk) => {
        size += chunk.length;
        if (size > MAX_DOWNLOAD) {
          res.destroy(new Error(`Download exceeds ${MAX_DOWNLOAD} bytes`));
          return;
        }
        hash.update(chunk);
      });
      res.on("error", (err) => {
        out.destroy();
        reject(err);
      });
      out.on("error", reject);
      out.on("finish", () => resolve(hash.digest("hex")));
      res.pipe(out);
    });
    req.setTimeout(TIMEOUT_MS, () => req.destroy(new Error("Download timed out")));
    req.on("error", reject);
  });
}

async function ensureBinary({ log = () => {} } = {}) {
  if (fs.existsSync(BIN_PATH)) {
    return BIN_PATH;
  }
  const name = assetName();
  const want = expectedHash(readChecksums(), name);

  fs.mkdirSync(BIN_DIR, { recursive: true });
  const tmp = fs.mkdtempSync(path.join(BIN_DIR, ".download-"));
  try {
    const archive = path.join(tmp, name);
    log(`Downloading spinnaker-mcp v${VERSION} for ${process.platform}-${process.arch}...`);
    const got = await download(`https://github.com/${REPO}/releases/download/v${VERSION}/${name}`, archive);
    if (got !== want) {
      const err = new Error(`Checksum mismatch for ${name}: expected ${want}, got ${got}`);
      err.code = "EINTEGRITY";
      throw err;
    }
    // No shell: fixed argv. bsdtar (macOS, Windows 10+) also reads the Windows zip.
    execFileSync("tar", ["-xf", archive, "-C", tmp, BIN_NAME], { stdio: "ignore" });
    const extracted = path.join(tmp, BIN_NAME);
    if (process.platform !== "win32") {
      fs.chmodSync(extracted, 0o755);
    }
    fs.renameSync(extracted, BIN_PATH);
    log(`Installed spinnaker-mcp to ${BIN_PATH} (sha256 verified)`);
    return BIN_PATH;
  } finally {
    fs.rmSync(tmp, { recursive: true, force: true });
  }
}

module.exports = { BIN_PATH, assetName, expectedHash, readChecksums, download, ensureBinary };
