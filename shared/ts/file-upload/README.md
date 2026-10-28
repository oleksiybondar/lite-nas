# Shared File Service upload

`@lite-nas/file-upload` exposes `UploadDialog` for LiteNAS React apps. Pass a
canonical absolute destination directory and an `onUploaded` callback. The
callback receives each completed file and may return a promise; the queue waits
for it before starting the next file. The
dialog accepts multiple selected or dropped files and uploads them one at a
time through File Service's `/f/proprietary/upload` protocol. File Service
authorizes create or replace at each operation; the dialog reports permission
errors from the service. Existing regular files may be replaced.
The checked-by-default "Readable by other users" option sends `private:false`;
unchecking it sends `private:true`. The service publishes shared files as
`0664` and private files as `0660`.

The browser sends the same-origin `lite-nas-at` cookie automatically. The
package does not handle or persist tokens. File content is hashed and sent in
bounded slices, so the whole file is never read into browser memory at once.

The upload request asks the service to fix a chunk size from the original file
size: up to 8 MiB uses 1 MiB chunks, up to 256 MiB uses 4 MiB, up to 2 GiB
uses 8 MiB, and larger files use 16 MiB. The last chunk may be shorter. These
are initial values within the File Service's 16 MiB maximum and can be tuned
later without changing the modal API.
