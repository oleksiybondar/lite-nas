import type { ReactNode } from "react";

/** Matches a literal search term against the text shown or described by a catalog item. */
export const matchesCatalogSearch = (
  search: string,
  values: readonly (string | number | undefined)[],
): boolean => {
  const needle = search.trim().toLocaleLowerCase();
  return (
    !needle ||
    values.some((value) =>
      String(value ?? "")
        .toLocaleLowerCase()
        .includes(needle),
    )
  );
};

/** Marks each literal match while keeping catalog text escaped by React. */
export const highlightCatalogText = (value: string | number, search: string): ReactNode => {
  const text = String(value);
  const needle = search.trim();
  if (!needle) return text;
  const lower = text.toLocaleLowerCase();
  const query = needle.toLocaleLowerCase();
  const parts: ReactNode[] = [];
  let start = 0;
  let match = lower.indexOf(query);
  while (match >= 0) {
    parts.push(text.slice(start, match));
    parts.push(<mark key={match}>{text.slice(match, match + needle.length)}</mark>);
    start = match + needle.length;
    match = lower.indexOf(query, start);
  }
  parts.push(text.slice(start));
  return parts;
};
