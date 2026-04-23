#!/usr/bin/env node
"use strict";

const { execFileSync } = require("child_process");
const crypto = require("crypto");
const fs = require("fs");
const path = require("path");
const https = require("https");

const VERSION = require("./package.json").version;
const REPO = "GeiserX/spinnaker-mcp";
const BIN_NAME = process.platform === "win32" ? "spinnaker-mcp.exe" : "spinnaker-mcp";
const BIN_DIR = path.join(__dirname, "bin");
const BIN_PATH = path.join(BIN_DIR, BIN_NAME);

function getPlatformArch() {
  const platform = process.platform;
  const arch = process.arch;

  const platformMap = { darwin: "darwin", linux: "linux", win32: "windows" };
  const archMap = { x64: "amd64", arm64: "arm64" };

  const p = platformMap[platform];
  const a = archMap[arch];

  if (!p || !a) {
    throw new Error(`Unsupported platform: ${platform}-${arch}`);
  }

  return { platform: p, arch: a };
}

function getAssetName() {
  const { platform, arch } = getPlatformArch();
  const ext = platform === "windows" ? "zip" : "tar.gz";
  return `spinnaker-mcp_${VERSION}_${platform}_${arch}.${ext}`;
}

function downloadFile(url, maxRedirects) {
  if (maxRedirects === undefined) maxRedirects = 5;
  if (maxRedirects <= 0) {
    return Promise.reject(new Error("Too many redirects"));
  }
  const parsed = new URL(url);
  if (parsed.protocol !== "https:") {
    return Promise.reject(new Error("Refusing non-HTTPS download URL: " + url));
  }
  return new Promise((resolve, reject) => {
    https.get(url, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        res.resume();
        const next = new URL(res.headers.location, url).toString();
        return downloadFile(next, maxRedirects - 1).then(resolve, reject);
      }
      if (res.statusCode !== 200) {
        res.resume();
        return reject(new Error(`Download failed: HTTP ${res.statusCode}`));
      }
      const chunks = [];
      res.on("data", (chunk) => chunks.push(chunk));
      res.on("end", () => resolve(Buffer.concat(chunks)));
      res.on("error", reject);
    }).on("error", reject);
  });
}

async function extract(buffer, assetName) {
  fs.mkdirSync(BIN_DIR, { recursive: true });

  if (assetName.endsWith(".zip")) {
    const tmpZip = path.join(BIN_DIR, "tmp.zip");
    fs.writeFileSync(tmpZip, buffer);
    execFileSync("unzip", ["-o", tmpZip, BIN_NAME, "-d", BIN_DIR], { stdio: "ignore" });
    fs.unlinkSync(tmpZip);
  } else {
    const tmpTar = path.join(BIN_DIR, "tmp.tar.gz");
    fs.writeFileSync(tmpTar, buffer);
    execFileSync("tar", ["-xzf", tmpTar, "-C", BIN_DIR, BIN_NAME], { stdio: "ignore" });
    fs.unlinkSync(tmpTar);
  }

  if (process.platform !== "win32") {
    fs.chmodSync(BIN_PATH, 0o755);
  }
}

async function main() {
  if (fs.existsSync(BIN_PATH)) {
    console.log("spinnaker-mcp binary already exists, skipping download.");
    return;
  }

  const assetName = getAssetName();
  const baseURL = `https://github.com/${REPO}/releases/download/v${VERSION}`;
  const assetURL = `${baseURL}/${assetName}`;
  const checksumURL = `${baseURL}/checksums.txt`;

  console.log(`Downloading spinnaker-mcp v${VERSION} for ${process.platform}-${process.arch}...`);
  const [buffer, checksumData] = await Promise.all([
    downloadFile(assetURL),
    downloadFile(checksumURL).catch(() => null),
  ]);

  if (checksumData) {
    const lines = checksumData.toString().split("\n");
    const line = lines.find((l) => l.includes(assetName));
    if (line) {
      const expectedHash = line.trim().split(/\s+/)[0];
      const actualHash = crypto.createHash("sha256").update(buffer).digest("hex");
      if (actualHash !== expectedHash) {
        throw new Error(
          `Checksum mismatch for ${assetName}: expected ${expectedHash}, got ${actualHash}`
        );
      }
      console.log("Checksum verified.");
    } else {
      console.warn(`Warning: no checksum entry found for ${assetName}, skipping verification.`);
    }
  } else {
    console.warn("Warning: could not download checksums.txt, skipping verification.");
  }

  console.log("Extracting binary...");
  await extract(buffer, assetName);

  console.log(`Installed spinnaker-mcp to ${BIN_PATH}`);
}

main().catch((err) => {
  console.error(`Failed to install spinnaker-mcp: ${err.message}`);
  process.exit(1);
});
