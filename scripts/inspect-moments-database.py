#!/usr/bin/env python3
"""Print non-destructive Moments gallery, media, scan, and location diagnostics."""

import sqlite3


def main() -> None:
    """Open the installed SQLite file read-only and print bounded diagnostics."""
    database = sqlite3.connect("file:/var/lib/lite-nas/moments/moments.db?mode=ro", uri=True)
    try:
        for label, query in (
            ("galleries", "SELECT count(*) FROM galleries"),
            ("media", "SELECT count(*) FROM gallery_media"),
            ("tag categories", "SELECT category,count(*) FROM tags GROUP BY category ORDER BY category"),
            ("locations", "SELECT value FROM tags WHERE category='LOCATION' ORDER BY value LIMIT 20"),
            ("scan statuses", "SELECT status,count(*) FROM media_scan_jobs GROUP BY status"),
        ):
            print(label, database.execute(query).fetchall())
    finally:
        database.close()


if __name__ == "__main__":
    main()
