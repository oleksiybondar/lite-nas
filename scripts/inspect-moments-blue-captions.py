#!/usr/bin/env python3
"""Print captions for Moments images found by a blue text match."""

import sqlite3


def main() -> None:
    """Read matching caption text without exposing physical file paths."""
    database = sqlite3.connect(
        "file:/var/lib/lite-nas/moments/moments.db?mode=ro", uri=True
    )
    try:
        rows = database.execute(
            """SELECT DISTINCT gm.path, mc.caption FROM gallery_media gm
               LEFT JOIN media_captions mc ON mc.path=gm.path
               WHERE gm.media_type IN ('IMAGE','LIVE_PHOTO') AND
                 (instr(COALESCE(mc.search_key,''),'blue')>0 OR EXISTS (
                   SELECT 1 FROM media_file_tags mt JOIN tags t ON t.id=mt.tag_id
                   WHERE mt.path=gm.path AND instr(t.search_key,'blue')>0))
               ORDER BY gm.path LIMIT 30"""
        )
        for index, (_, caption) in enumerate(rows, start=1):
            print(index, repr(caption))
    finally:
        database.close()


if __name__ == "__main__":
    main()
