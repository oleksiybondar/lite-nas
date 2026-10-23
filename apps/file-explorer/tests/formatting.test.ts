import { formatByteSize } from "@lite-nas/formatting";
import { expect, test } from "vitest";

test.each([
  ["unknown", null, "—"],
  ["zero bytes", 0, "0 B"],
  ["largest byte count", 1023, "1023 B"],
  ["kilobyte", 1024, "1 KB"],
  ["fractional megabyte", 1.5 * 1024 ** 2, "1.5 MB"],
  ["gigabyte", 1024 ** 3, "1 GB"],
  ["terabyte", 1024 ** 4, "1 TB"],
  ["unit ceiling", 1024 ** 5, "1024 TB"],
])("file-explorer-app/FR-003 formats %s", (_case, bytes, expected) => {
  expect(formatByteSize(bytes)).toBe(expected);
});
