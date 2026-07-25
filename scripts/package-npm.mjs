#!/usr/bin/env node

import { chmod, cp, mkdir, readFile, writeFile } from "node:fs/promises";
import { spawnSync } from "node:child_process";
import path from "node:path";
import process from "node:process";

const version = process.argv[2];
if (!version || !/^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$/.test(version)) {
  process.stderr.write("usage: node scripts/package-npm.mjs VERSION\n");
  process.exit(2);
}

const targets = [
  ["darwin", "arm64", "darwin", "arm64"],
  ["darwin", "x64", "darwin", "amd64"],
  ["linux", "arm64", "linux", "arm64"],
  ["linux", "x64", "linux", "amd64"],
  ["win32", "arm64", "windows", "arm64"],
  ["win32", "x64", "windows", "amd64"],
];

const outputRoot = path.resolve("dist/npm");
await mkdir(outputRoot, { recursive: true });

for (const [npmOS, npmCPU, goOS, goArch] of targets) {
  const packageName = `pb-agent-${npmOS}-${npmCPU}`;
  const packageDir = path.join(outputRoot, packageName);
  const binDir = path.join(packageDir, "bin");
  const binaryName = `pb-agent${goOS === "windows" ? ".exe" : ""}`;
  const binaryPath = path.join(binDir, binaryName);
  await mkdir(binDir, { recursive: true });

  const build = spawnSync(
    "go",
    [
      "build",
      "-trimpath",
      `-ldflags=-s -w -X github.com/anirudh-777/pb-agent/internal/version.Version=${version}`,
      "-o",
      binaryPath,
      "./cmd/pb-agent",
    ],
    {
      stdio: "inherit",
      env: { ...process.env, CGO_ENABLED: "0", GOOS: goOS, GOARCH: goArch },
    },
  );
  if (build.status !== 0) process.exit(build.status ?? 1);
  if (goOS !== "windows") await chmod(binaryPath, 0o755);

  await writeFile(
    path.join(packageDir, "package.json"),
    JSON.stringify(
      {
        name: packageName,
        version,
        description: `Native ${npmOS} ${npmCPU} binary for pb-agent`,
        license: "Apache-2.0",
        repository: "github:anirudh-777/pb-agent",
        os: [npmOS],
        cpu: [npmCPU],
        files: ["bin"],
      },
      null,
      2,
    ) + "\n",
  );
  await cp("LICENSE", path.join(packageDir, "LICENSE"));
}

const metaDir = path.join(outputRoot, "pb-agent");
await mkdir(path.join(metaDir, "bin"), { recursive: true });
const meta = JSON.parse(await readFile("npm/pb-agent/package.json", "utf8"));
meta.version = version;
for (const dependency of Object.keys(meta.optionalDependencies)) {
  meta.optionalDependencies[dependency] = version;
}
await writeFile(path.join(metaDir, "package.json"), JSON.stringify(meta, null, 2) + "\n");
await cp("npm/pb-agent/bin/pb-agent.js", path.join(metaDir, "bin/pb-agent.js"));
await chmod(path.join(metaDir, "bin/pb-agent.js"), 0o755);
await cp("README.md", path.join(metaDir, "README.md"));
await cp("LICENSE", path.join(metaDir, "LICENSE"));
