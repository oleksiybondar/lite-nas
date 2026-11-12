/** A source line paired with its position and stable duplicate-aware React key. */
export type KeyedLine = {
  line: string;
  index: number;
  key: string;
};

/** Assigns stable content and occurrence keys to lines in one linear pass. */
export const keyLines = (lines: string[]): KeyedLine[] => {
  const occurrences = new Map<string, number>();
  return lines.map((line, index) => {
    const occurrence = occurrences.get(line) ?? 0;
    occurrences.set(line, occurrence + 1);
    return { line, index, key: `${line}:${occurrence}` };
  });
};
