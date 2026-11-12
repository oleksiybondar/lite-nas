/** Builds one File Service listing object for browser workflow tests. */
export function listingEntry(path: string, kind: "file" | "directory") {
  return {
    path,
    name: path.slice(path.lastIndexOf("/") + 1),
    kind,
    size: null,
    owner: null,
    group: null,
    permissions: null,
    last_modified: null,
  };
}
