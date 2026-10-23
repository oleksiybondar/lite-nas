import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { expect, test } from "vitest";
import { highlightSource } from "../src/highlight-source";

test.each([
  ["TypeScript", "ts", "const answer: number = 42;"],
  ["Python", "py", "def answer(): return 42"],
  ["JSON", "json", '{"answer": 42}'],
  ["Markdown", "md", "# File preview"],
] as const)("file-explorer-app/FR-003 highlights %s source", (_name, extension, source) => {
  const markup = renderToStaticMarkup(
    createElement("code", null, highlightSource(source, extension)),
  );
  expect({ hasTokenMarkup: markup.includes("token "), hasOutput: markup.length > 0 }).toEqual({
    hasTokenMarkup: true,
    hasOutput: true,
  });
});

test("file-explorer-app/FR-003 escapes markup in highlighted source", () => {
  const markup = renderToStaticMarkup(
    createElement("code", null, highlightSource("<script>alert('unsafe')</script>", "html")),
  );
  expect({
    containsEscapedMarkup: markup.includes("&lt;") && markup.includes("&gt;"),
    containsExecutableScript: markup.includes("<script>"),
  }).toEqual({
    containsEscapedMarkup: true,
    containsExecutableScript: false,
  });
});

test("file-explorer-app/FR-003 leaves plain text unhighlighted", () => {
  expect(highlightSource("plain words", "txt")).toBeNull();
});
