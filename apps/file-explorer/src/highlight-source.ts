import { createElement, type ReactNode } from "react";
import Prism, { loadLanguages, type TokenizedSource } from "reprism";
import bash from "reprism/languages/bash";
import c from "reprism/languages/c";
import cpp from "reprism/languages/cpp";
import diff from "reprism/languages/diff";
import go from "reprism/languages/go";
import ini from "reprism/languages/ini";
import java from "reprism/languages/java";
import json from "reprism/languages/json";
import jsx from "reprism/languages/jsx";
import markdown from "reprism/languages/markdown";
import python from "reprism/languages/python";
import rust from "reprism/languages/rust";
import typescript from "reprism/languages/typescript";
import yaml from "reprism/languages/yaml";

loadLanguages([
  bash,
  c,
  cpp,
  diff,
  go,
  ini,
  java,
  json,
  jsx,
  markdown,
  python,
  rust,
  typescript,
  yaml,
]);

const sourceLanguageByExtension: Record<string, string> = {
  bash: "bash",
  c: "c",
  cjs: "javascript",
  cpp: "cpp",
  css: "css",
  diff: "diff",
  go: "go",
  h: "c",
  htm: "markup",
  html: "markup",
  hpp: "cpp",
  ini: "ini",
  java: "java",
  js: "javascript",
  json: "json",
  jsx: "jsx",
  md: "markdown",
  mjs: "javascript",
  py: "python",
  rs: "rust",
  sh: "bash",
  ts: "typescript",
  tsx: "jsx",
  xml: "markup",
  yaml: "yaml",
  yml: "yaml",
};

/** Converts a parsed token into React text or a syntax-classed span. */
function renderToken(token: TokenizedSource, key: number): ReactNode {
  if (typeof token === "string") return token;
  const aliases: string[] = [];
  if (Array.isArray(token.alias)) aliases.push(...token.alias);
  else if (token.alias !== undefined) aliases.push(token.alias);
  const content = Array.isArray(token.content)
    ? token.content.map((child, index) => renderToken(child, index))
    : renderToken(token.content, 0);
  const className = ["token", token.type, ...aliases].join(" ");
  return createElement("span", { className, key }, content);
}

/** Returns safely rendered syntax tokens for a recognized source extension. */
export function highlightSource(source: string, extension: string): ReactNode | null {
  const language = sourceLanguageByExtension[extension.toLowerCase()];
  if (language === undefined) return null;
  return Prism.tokenize(source, Prism.languages[language]).map((token, index) =>
    renderToken(token, index),
  );
}
