"""JSON Schemas for live responses from the installed REST services."""

AUTH_ME_SCHEMA: dict[str, object] = {
    "type": "object",
    "required": ["success", "timestamp", "data"],
    "properties": {
        "success": {"const": True},
        "timestamp": {"type": "string"},
        "data": {
            "type": "object",
            "required": ["authenticated", "auth_type", "user", "roles", "scopes"],
            "properties": {
                "authenticated": {"const": True},
                "auth_type": {"type": "string"},
                "user": {
                    "type": "object",
                    "required": ["id", "login"],
                    "properties": {"id": {"type": "string"}, "login": {"type": "string"}},
                },
                "roles": {"type": "array", "items": {"type": "string"}},
                "scopes": {"type": "array", "items": {"type": "string"}},
            },
        },
    },
}

FILE_STAT_SCHEMA: dict[str, object] = {
    "type": "object",
    "required": ["path", "kind"],
    "properties": {
        "path": {"type": "string", "minLength": 1},
        "kind": {"enum": ["file", "directory"]},
    },
    "additionalProperties": False,
}

GALLERY_SCHEMA: dict[str, object] = {
    "type": "object",
    "required": [
        "id",
        "type",
        "name",
        "description",
        "ownerId",
        "visibility",
        "uploadDirectory",
        "wallRenderStrategy",
        "thumbnailImageUrl",
        "tags",
        "createdAt",
        "updatedAt",
    ],
    "properties": {
        "id": {"type": "string", "format": "uuid"},
        "type": {"enum": ["ALBUM", "COLLECTION"]},
        "name": {"type": "string", "minLength": 1},
        "description": {"type": "string"},
        "ownerId": {"type": "string", "minLength": 1},
        "visibility": {"enum": ["PRIVATE", "PUBLIC"]},
        "uploadDirectory": {"type": ["string", "null"]},
        "wallRenderStrategy": {"enum": ["mosaic", "justified", "masonry", "grid"]},
        "thumbnailImageUrl": {"type": ["string", "null"]},
        "tags": {"type": "array"},
        "createdAt": {"type": "string"},
        "updatedAt": {"type": "string"},
    },
}

GALLERY_LIST_SCHEMA: dict[str, object] = {
    "type": "object",
    "required": ["items", "metadata"],
    "properties": {
        "items": {"type": "array", "items": GALLERY_SCHEMA},
        "metadata": {
            "type": "object",
            "required": ["page", "size", "total_count", "total_pages"],
            "properties": {
                "page": {"type": "integer", "minimum": 1},
                "size": {"type": "integer", "minimum": 1},
                "total_count": {"type": "integer", "minimum": 0},
                "total_pages": {"type": "integer", "minimum": 0},
            },
        },
    },
}

STREAM_OPENAPI_SCHEMA: dict[str, object] = {
    "type": "object",
    "required": ["openapi", "info", "paths"],
    "properties": {
        "openapi": {"type": "string", "pattern": "^3\\."},
        "info": {
            "type": "object",
            "required": ["title", "version"],
            "properties": {
                "title": {"type": "string", "minLength": 1},
                "version": {"type": "string", "minLength": 1},
            },
        },
        "paths": {
            "type": "object",
            "required": [
                "/s/{absolutePath}",
                "/s/preview/{absolutePath}",
                "/s/transcode/{absolutePath}",
            ],
        },
    },
}
