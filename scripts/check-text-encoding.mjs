import { execFileSync } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { TextDecoder } from "node:util";

const textExtensions = new Set([
  ".bat",
  ".css",
  ".go",
  ".html",
  ".js",
  ".json",
  ".md",
  ".mjs",
  ".ps1",
  ".ts",
  ".tsx",
  ".yaml",
  ".yml",
]);

const ignoredPaths = new Set([
  "specifications/gameworld_data_model.go",
]);

const checkedExactPaths = new Set([
  ".editorconfig",
  ".gitattributes",
  ".githooks/pre-commit",
]);

const mojibakePattern = new RegExp(
  String.raw`(?:\u0420[\u00a0\u00ab\u00b0\u00b1\u00bb\u0402\u0403\u040e\u0409\u040a\u040b\u040c\u040f\u0412\u0459\u045a\u045f\u2018\u2019\u201a\u201c\u201d\u201e\u2020\u2021\u2022\u2026\u2013\u2014\u2030\u2039\u20ac\u2122]|\u0421[\u0402\u0403\u0409\u040a\u040b\u040c\u040f\u0459\u045a\u045f\u2018\u2019\u201a\u201c\u201d\u201e\u2020\u2021\u2022\u2026\u2013\u2014\u2030\u2039\u20ac\u2122]|\u0432\u0402|\u0420\u0406\u0420|\u0420\u040e\u0420|\u0420\u040b\u0420|\ufffd)`,
  "u",
);
const utf8Decoder = new TextDecoder("utf-8", { fatal: true });

const runGit = (args) => execFileSync("git", args, { encoding: "utf8" });

const normalizePath = (filePath) => filePath.replaceAll("\\", "/");

const isCheckedTextPath = (filePath) => {
  const normalizedPath = normalizePath(filePath);

  return !ignoredPaths.has(normalizedPath)
    && (checkedExactPaths.has(normalizedPath) || textExtensions.has(path.extname(normalizedPath).toLowerCase()));
};

const stagedMode = process.argv.includes("--staged");

const listFiles = () => {
  const output = stagedMode
    ? runGit(["diff", "--cached", "--name-only", "--diff-filter=ACMR"])
    : [
      runGit(["ls-files"]),
      runGit(["ls-files", "--others", "--exclude-standard"]),
    ].join("\n");

  return output
    .split(/\r?\n/)
    .map((filePath) => filePath.trim())
    .filter(Boolean)
    .filter(isCheckedTextPath);
};

const readFileBytes = (filePath) => {
  if (stagedMode) {
    return execFileSync("git", ["show", `:${normalizePath(filePath)}`]);
  }

  return readFileSync(filePath);
};

const formatLine = (line) => line.trim().replace(/\s+/g, " ").slice(0, 160);

const checkFile = (filePath) => {
  if (!stagedMode && !existsSync(filePath)) {
    return [];
  }

  const errors = [];
  let text = "";

  try {
    text = utf8Decoder.decode(readFileBytes(filePath));
  } catch {
    return [`${filePath}: файл не читается как UTF-8`];
  }

  text.split(/\r?\n/).forEach((line, index) => {
    if (mojibakePattern.test(line)) {
      errors.push(`${filePath}:${index + 1}: похоже на кракозябры: ${formatLine(line)}`);
    }
  });

  return errors;
};

const errors = listFiles().flatMap(checkFile);

if (errors.length > 0) {
  console.error("Найдены проблемы с кодировкой текста:");
  for (const error of errors.slice(0, 80)) {
    console.error(error);
  }
  if (errors.length > 80) {
    console.error(`...и ещё ${errors.length - 80}`);
  }
  process.exit(1);
}
