import ArchiveIcon from "@mui/icons-material/Archive";
import DescriptionIcon from "@mui/icons-material/Description";
import FolderIcon from "@mui/icons-material/Folder";
import ImageIcon from "@mui/icons-material/Image";
import InsertDriveFileIcon from "@mui/icons-material/InsertDriveFile";
import MovieIcon from "@mui/icons-material/Movie";
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
  json: <DescriptionIcon color="info" />,
  xml: <DescriptionIcon color="info" />,
  yaml: <DescriptionIcon color="info" />,
  yml: <DescriptionIcon color="info" />,
  log: <DescriptionIcon color="info" />,
  js: <DescriptionIcon color="info" />,
  ts: <DescriptionIcon color="info" />,
  py: <DescriptionIcon color="info" />,
  go: <DescriptionIcon color="info" />,
  sh: <DescriptionIcon color="info" />,
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
