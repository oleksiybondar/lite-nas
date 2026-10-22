# First-version raw reader review

The first File Service version reads an RBAC-authorized path by invoking the
fixed `/bin/cat` executable through noninteractive `sudo -u` as the verified
Unix user. This lets Unix deny a read the user cannot perform, including reads
of another user's `0600` file. The service must still validate the access
token, resolve the RBAC UID, and obtain a fresh RBAC read allow before every
request. Sudo and Unix permissions do not replace those checks.

## Known limitation

RBAC checks a pathname before `cat` opens it. Another process with write
access to the parent directory can replace the pathname between those events.
`cat` cannot report or verify the identity of the inode it opened. A request
can therefore return bytes from a different file even after a correct RBAC
decision. The caller has explicitly accepted this race for the first version;
the task's strict race-resistant read requirement is deferred for this path.

## Required review before hardening

1. Add a controlled test that swaps the pathname after RBAC allows it and
   before the reader opens it. Record whether bytes from the replacement are
   exposed on both `/f/` and `/d/`, including resumed range requests.
2. Review sudoers and AppArmor together: a broad `cat` grant must not become a
   shell grant, root target-user grant, or access to paths outside configured
   storage roots.
3. Replace raw `cat` with a fixed reader that opens a canonical path without
   following a substituted symlink, verifies device and inode against the
   authorized object, and streams a requested byte range from that open file.
4. Repeat the replacement test and verify denial when the inode changes.
5. Test a file changed to `0600` after provisioning and a directory changed to
   `0700` after provisioning. A file should be readable through user-scoped
   `cat` when its ancestors remain traversable; a private directory whose ACL
   mask removes the read group's search permission currently prevents the
   service's path-resolution and stat checks before `cat` starts. Define a
   user-scoped, race-checked metadata lookup before claiming support for that
   case.

This review is required before claiming that downloads are race-resistant or
that pathname changes cannot bypass an RBAC decision.
