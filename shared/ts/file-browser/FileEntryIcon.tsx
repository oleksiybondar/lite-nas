import ArchiveIcon from "@mui/icons-material/Archive";
import CodeIcon from "@mui/icons-material/Code";
import CssIcon from "@mui/icons-material/Css";
import DataObjectIcon from "@mui/icons-material/DataObject";
import DescriptionIcon from "@mui/icons-material/Description";
import FolderIcon from "@mui/icons-material/Folder";
import HtmlIcon from "@mui/icons-material/Html";
import ImageIcon from "@mui/icons-material/Image";
import InsertDriveFileIcon from "@mui/icons-material/InsertDriveFile";
import JavascriptIcon from "@mui/icons-material/Javascript";
import MovieIcon from "@mui/icons-material/Movie";
import MusicNoteIcon from "@mui/icons-material/MusicNote";
import PictureAsPdfIcon from "@mui/icons-material/PictureAsPdf";
import type { ReactElement } from "react";

const fileIcons: Record<string, ReactElement> = {
  pdf: <PictureAsPdfIcon color="error" />,
  mp4: <MovieIcon color="secondary" />,
  m4v: <MovieIcon color="secondary" />,
  mov: <MovieIcon color="secondary" />,
  avi: <MovieIcon color="secondary" />,
  mkv: <MovieIcon color="secondary" />,
  webm: <MovieIcon color="secondary" />,
  ogv: <MovieIcon color="secondary" />,
  mp3: <MusicNoteIcon color="secondary" />,
  wav: <MusicNoteIcon color="secondary" />,
  m4a: <MusicNoteIcon color="secondary" />,
  aac: <MusicNoteIcon color="secondary" />,
  flac: <MusicNoteIcon color="secondary" />,
  ogg: <MusicNoteIcon color="secondary" />,
  jpg: <ImageIcon color="success" />,
  jpeg: <ImageIcon color="success" />,
  png: <ImageIcon color="success" />,
  gif: <ImageIcon color="success" />,
  webp: <ImageIcon color="success" />,
  avif: <ImageIcon color="success" />,
  bmp: <ImageIcon color="success" />,
  svg: <ImageIcon color="success" />,
  heic: <ImageIcon color="success" />,
  zip: <ArchiveIcon color="warning" />,
  "7z": <ArchiveIcon color="warning" />,
  rar: <ArchiveIcon color="warning" />,
  tar: <ArchiveIcon color="warning" />,
  gz: <ArchiveIcon color="warning" />,
  bz2: <ArchiveIcon color="warning" />,
  xz: <ArchiveIcon color="warning" />,
  zst: <ArchiveIcon color="warning" />,
  txt: <DescriptionIcon color="info" />,
  md: <DescriptionIcon color="info" />,
  csv: <DescriptionIcon color="info" />,
  json: <DataObjectIcon color="info" />,
  xml: <DescriptionIcon color="info" />,
  yaml: <DescriptionIcon color="info" />,
  yml: <DescriptionIcon color="info" />,
  log: <DescriptionIcon color="info" />,
  html: <HtmlIcon color="info" />,
  htm: <HtmlIcon color="info" />,
  css: <CssIcon color="info" />,
  js: <JavascriptIcon color="info" />,
  jsx: <JavascriptIcon color="info" />,
  mjs: <JavascriptIcon color="info" />,
  cjs: <JavascriptIcon color="info" />,
  ts: <CodeIcon color="info" />,
  tsx: <CodeIcon color="info" />,
  py: <CodeIcon color="info" />,
  go: <CodeIcon color="info" />,
  rs: <CodeIcon color="info" />,
  java: <CodeIcon color="info" />,
  c: <CodeIcon color="info" />,
  h: <CodeIcon color="info" />,
  cpp: <CodeIcon color="info" />,
  hpp: <CodeIcon color="info" />,
  sh: <CodeIcon color="info" />,
  bash: <CodeIcon color="info" />,
  sql: <DataObjectIcon color="info" />,
};

/** Recognizable file category icon selected from a File Service path. */
export function FileEntryIcon({
  kind,
  name,
}: {
  kind: "directory" | "file";
  name: string;
}): ReactElement {
  if (kind === "directory") {
    return <FolderIcon color="primary" />;
  }
  const extension = name.includes(".") ? name.toLowerCase().split(".").pop() : undefined;
  return fileIcons[extension ?? ""] ?? <InsertDriveFileIcon color="action" />;
}
