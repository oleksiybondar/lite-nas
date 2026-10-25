# Shared file browser

`@lite-nas/file-browser` exposes the validated File Service client, path and
response schemas, explorer state, directory panel, and reusable folder and file
selection dialogs.

```tsx
<FileSelectionDialog
  open={fileDialogOpen}
  onClose={() => setFileDialogOpen(false)}
  initialDirectory="/home/oleksiybondar"
  extensionFilter={{ include: ["jpg", "jpeg", "png"], exclude: ["png"] }}
  onSelect={(path) => setSelectedPath(path)}
/>
```

Use `FolderSelectionDialog` for folder picking. `FileSelectionDialog` is the file
picker and accepts an optional extension filter.

`include` and `exclude` are optional, case-insensitive extension lists. Entries
may use `jpg` or `.jpg`. When both are set, `exclude` wins. Folders remain visible
in the file picker for navigation. The folder picker shows folders only. Both
pickers hide dot-prefixed entries by default and provide a switch to show them.

The selection callback receives a validated absolute host path. It does not
perform a File Service mutation; the embedding application decides what to do
with the chosen path.
