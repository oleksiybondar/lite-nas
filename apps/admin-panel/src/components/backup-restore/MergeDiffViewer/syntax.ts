import { createElement, type ReactElement, type ReactNode } from "react";
import { highlight } from "reprism";

/** Selects one supported syntax grammar from the configuration file extension. */
export const syntaxForPath = (path: string): string | null => {
  const extension = path.split(".").pop()?.toLowerCase();
  if (extension === "yaml" || extension === "yml") return "yaml";
  if (extension === "json") return "json";
  if (extension === "ini") return "ini";
  return null;
};

/** Highlights one source line or safely escapes it when no grammar is available. */
export const highlightCodeLine = (line: string, language: string | null): ReactNode[] => {
  const source = line || " ";
  if (!language) return [source];

  return parseHighlightedMarkup(highlight(source, language, { component: "" }));
};

/** Converts reprism's escaped text and token spans into React-owned nodes. */
const parseHighlightedMarkup = (markup: string): ReactNode[] => {
  const output: ReactNode[] = [];
  const stack: Array<{ token: string; children: ReactNode[] }> = [];
  const tokenMarkup = /<span class="token ([\w-]+)">|<\/span>/g;
  let offset = 0;
  let match = tokenMarkup.exec(markup);
  while (match) {
    appendText(stack, output, markup.slice(offset, match.index));
    if (match[1]) stack.push({ token: match[1], children: [] });
    else appendClosedToken(stack, output, match.index);
    offset = tokenMarkup.lastIndex;
    match = tokenMarkup.exec(markup);
  }
  appendText(stack, output, markup.slice(offset));
  return output;
};

/** Adds a decoded highlighter text segment to the active token or root. */
const appendText = (
  stack: Array<{ token: string; children: ReactNode[] }>,
  output: ReactNode[],
  rawText: string,
): void => {
  const text = decodeHtml(rawText);
  if (text) (stack.at(-1)?.children ?? output).push(text);
};

/** Closes the current token span and attaches it to its parent token. */
const appendClosedToken = (
  stack: Array<{ token: string; children: ReactNode[] }>,
  output: ReactNode[],
  key: number,
): void => {
  const token = stack.pop();
  if (!token) return;
  const element: ReactElement = createElement(
    "span",
    { className: `token ${token.token}`, key: `${key}-${token.token}` },
    token.children,
  );
  (stack.at(-1)?.children ?? output).push(element);
};

/** Decodes only the entities emitted by reprism after React has taken over text rendering. */
const decodeHtml = (value: string): string =>
  value
    .replaceAll("&lt;", "<")
    .replaceAll("&gt;", ">")
    .replaceAll("&amp;", "&")
    .replaceAll("&quot;", '"');
