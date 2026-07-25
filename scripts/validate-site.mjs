#!/usr/bin/env node

import { existsSync, readdirSync, readFileSync } from "node:fs";
import path from "node:path";

const root = path.resolve("site");
const pages = [];
const errors = [];
const titles = new Map();

function walk(directory) {
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    const file = path.join(directory, entry.name);
    if (entry.isDirectory()) walk(file);
    if (entry.isFile() && entry.name.endsWith(".html")) pages.push(file);
  }
}

function report(file, message) {
  errors.push(`${path.relative(root, file)}: ${message}`);
}

walk(root);

for (const file of pages) {
  const html = readFileSync(file, "utf8");
  const title = html.match(/<title>(.*?)<\/title>/)?.[1];
  const description = html.match(/<meta name="description" content="([^"]+)"/)?.[1];
  const h1Count = [...html.matchAll(/<h1(?:\s[^>]*)?>/g)].length;

  if (!title) report(file, "missing title");
  if (title && titles.has(title)) report(file, `duplicate title also used by ${titles.get(title)}`);
  if (title) titles.set(title, path.relative(root, file));
  if (!file.endsWith("404.html") && !description) report(file, "missing meta description");
  if (!file.endsWith("404.html") && !/<link rel="canonical"/.test(html)) report(file, "missing canonical URL");
  if (h1Count !== 1) report(file, `expected one h1, found ${h1Count}`);

  for (const match of html.matchAll(/(?:href|src)="(\/pb-agent\/[^"]*)"/g)) {
    const urlPath = match[1].split(/[?#]/)[0].replace(/^\/pb-agent\/?/, "");
    let target = path.join(root, urlPath);
    if (urlPath === "" || urlPath.endsWith("/")) target = path.join(target, "index.html");
    if (!existsSync(target)) report(file, `broken internal link ${match[1]}`);
  }

  for (const match of html.matchAll(/<script type="application\/ld\+json">([\s\S]*?)<\/script>/g)) {
    try {
      JSON.parse(match[1]);
    } catch (error) {
      report(file, `invalid JSON-LD: ${error.message}`);
    }
  }
}

for (const required of ["robots.txt", "sitemap.xml", "llms.txt", "pricing.md", "og.png", ".nojekyll"]) {
  if (!existsSync(path.join(root, required))) errors.push(`missing required site file: ${required}`);
}

if (errors.length > 0) {
  process.stderr.write(`${errors.join("\n")}\n`);
  process.exit(1);
}

process.stdout.write(`Validated ${pages.length} HTML pages and required SEO/GEO files.\n`);
