import { afterEach, expect, test, vi } from "vitest";
import {
  addCollectionDirectory,
  cancelGalleryScan,
  changeCollectionFile,
  chooseGalleryThumbnailPath,
  createGallery,
  galleryFiles,
  getGalleryScan,
  latestGalleryScan,
  listGalleries,
  removeCollectionDirectory,
  saveGalleryThumbnail,
  shareGalleryByUsername,
  startGalleryScan,
  updateGallery,
} from "./moments-data";
import type { GalleryRecord } from "./schemas";

const gallery: GalleryRecord = {
  id: "00000000-0000-4000-8000-000000000001",
  type: "ALBUM",
  name: "Photos",
  description: "",
  ownerId: "alice",
  visibility: "PRIVATE",
  uploadDirectory: "/NASPool/Photos",
  wallRenderStrategy: "masonry",
  thumbnailImageUrl: null,
  thumbnailPath: null,
  createdAt: "2026-10-26T00:00:00Z",
  updatedAt: "2026-10-26T00:00:00Z",
};

const scan = {
  id: "00000000-0000-4000-8000-000000000002",
  galleryId: gallery.id,
  kind: "MEDIA",
  status: "PENDING",
  totalItems: null,
  processedItems: 0,
  successfulItems: 0,
  failedItems: 0,
  createdAt: "2026-10-26T00:00:00Z",
  startedAt: null,
  finishedAt: null,
  errorSummary: null,
};

afterEach(() => vi.unstubAllGlobals());

test("media scan start uses the owner gallery route", async () => {
  const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify(scan), { status: 202 }));
  vi.stubGlobal("fetch", fetchMock);
  await startGalleryScan(gallery.id);
  expect({ path: fetchMock.mock.calls[0][0], method: fetchMock.mock.calls[0][1].method }).toEqual({
    path: `/api/galleries/${gallery.id}/scans`,
    method: "POST",
  });
});

test("media scan history reads the latest owner job", async () => {
  const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ items: [scan] })));
  vi.stubGlobal("fetch", fetchMock);
  await latestGalleryScan(gallery.id);
  expect(fetchMock.mock.calls[0][0]).toBe(`/api/galleries/${gallery.id}/scans?limit=1&kind=MEDIA`);
});

test("caption scan history reads only caption jobs", async () => {
  const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ items: [] })));
  vi.stubGlobal("fetch", fetchMock);
  await latestGalleryScan(gallery.id, "CAPTION");
  expect(fetchMock.mock.calls[0][0]).toBe(
    `/api/galleries/${gallery.id}/scans?limit=1&kind=CAPTION`,
  );
});

test("media scan polling reads a validated job", async () => {
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify(scan))));
  await expect(getGalleryScan(scan.id)).resolves.toMatchObject({ id: scan.id, status: "PENDING" });
});

test("media scan cancellation uses the same-user job route", async () => {
  const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify(scan)));
  vi.stubGlobal("fetch", fetchMock);
  await cancelGalleryScan(scan.id);
  expect({ path: fetchMock.mock.calls[0][0], method: fetchMock.mock.calls[0][1].method }).toEqual({
    path: `/api/scans/${scan.id}`,
    method: "DELETE",
  });
});

test("gallery feed requests the owner's paged backend results", async () => {
  const fetchMock = vi.fn().mockResolvedValue(
    new Response(
      JSON.stringify({
        items: [gallery],
        metadata: { page: 2, size: 20, total_count: 21, total_pages: 2 },
      }),
      { status: 200 },
    ),
  );
  vi.stubGlobal("fetch", fetchMock);
  await listGalleries(2, "alice");
  expect(fetchMock.mock.calls[0][0]).toBe("/api/galleries?page=2&size=20&ownerId=alice");
});

test("gallery search validates linked galleries on an image match", async () => {
  const link = { id: gallery.id, name: gallery.name, type: gallery.type, ownerId: gallery.ownerId };
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          items: [gallery],
          metadata: { page: 1, size: 20, total_count: 1, total_pages: 1 },
          matchingMedia: {
            items: [
              {
                galleryId: gallery.id,
                linkedGalleries: [link],
                path: "/NASPool/Photos/dog.jpg",
                mediaType: "IMAGE",
                width: 640,
                height: 480,
                livePhotoPath: null,
                caption: "A dog on a sofa",
                tags: [],
              },
            ],
            total: 1,
            limit: 50,
            offset: 0,
          },
        }),
      ),
    ),
  );
  const result = await listGalleries(1, undefined, { q: "dog" });
  expect(result.matchingMedia?.items[0]?.linkedGalleries).toEqual([link]);
});

test("image results use their own pagination offset", async () => {
  const fetchMock = vi.fn().mockResolvedValue(
    new Response(
      JSON.stringify({
        items: [],
        metadata: { page: 1, size: 20, total_count: 0, total_pages: 0 },
      }),
    ),
  );
  vi.stubGlobal("fetch", fetchMock);
  await listGalleries(1, undefined, { q: "dog" }, 50);
  expect(fetchMock.mock.calls[0][0]).toBe("/api/galleries?page=1&size=20&mediaOffset=50&q=dog");
});

test("gallery search sends supported filters to the backend", async () => {
  const fetchMock = vi.fn().mockResolvedValue(
    new Response(
      JSON.stringify({
        items: [],
        metadata: { page: 1, size: 20, total_count: 0, total_pages: 0 },
      }),
      { status: 200 },
    ),
  );
  vi.stubGlobal("fetch", fetchMock);
  await listGalleries(1, undefined, { q: "beach", type: "ALBUM", visibility: "PUBLIC" });
  expect(fetchMock.mock.calls[0][0]).toBe(
    "/api/galleries?page=1&size=20&q=beach&type=ALBUM&visibility=PUBLIC",
  );
});

test("gallery search sends media choices as repeated query values", async () => {
  const fetchMock = vi.fn().mockResolvedValue(
    new Response(
      JSON.stringify({
        items: [],
        metadata: { page: 1, size: 20, total_count: 0, total_pages: 0 },
      }),
      { status: 200 },
    ),
  );
  vi.stubGlobal("fetch", fetchMock);
  await listGalleries(1, undefined, { mediaType: ["IMAGE", "VIDEO"] });
  expect(fetchMock.mock.calls[0][0]).toBe(
    "/api/galleries?page=1&size=20&mediaType=IMAGE&mediaType=VIDEO",
  );
});

test("gallery search rejects duplicate media choices before transport", async () => {
  vi.stubGlobal("fetch", vi.fn());
  await expect(listGalleries(1, undefined, { mediaType: ["IMAGE", "IMAGE"] })).rejects.toThrow();
});

test("album media comes from its configured File Service directory", async () => {
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify([
          {
            path: "/NASPool/Photos/a.jpg",
            name: "a.jpg",
            kind: "file",
            size: null,
            owner: null,
            group: null,
            permissions: null,
            last_modified: null,
          },
        ]),
        { status: 200 },
      ),
    ),
  );
  await expect(galleryFiles(gallery, new AbortController().signal)).resolves.toEqual([
    "/NASPool/Photos/a.jpg",
  ]);
});

test("creating an album validates its source directory before transport", async () => {
  const fetchMock = vi.fn();
  vi.stubGlobal("fetch", fetchMock);
  await expect(
    createGallery({ type: "ALBUM", name: "Photos", uploadDirectory: "relative" }),
  ).rejects.toThrow();
});

test("adding media uses the collection file-reference route", async () => {
  const fetchMock = vi.fn().mockResolvedValue(new Response("{}", { status: 200 }));
  vi.stubGlobal("fetch", fetchMock);
  await changeCollectionFile(gallery.id, "/NASPool/Photos/a.jpg", true);
  expect({ path: fetchMock.mock.calls[0][0], method: fetchMock.mock.calls[0][1].method }).toEqual({
    path: `/api/collections/${gallery.id}/files`,
    method: "POST",
  });
});

test("editing a gallery validates metadata before the patch request", async () => {
  const fetchMock = vi.fn();
  vi.stubGlobal("fetch", fetchMock);
  await expect(
    updateGallery(gallery.id, {
      name: "Photos",
      description: "x".repeat(4001),
      visibility: "PRIVATE",
    }),
  ).rejects.toThrow();
});

test("owner layout changes are sent to the gallery metadata API", async () => {
  const fetchMock = vi
    .fn()
    .mockResolvedValue(new Response(JSON.stringify(gallery), { status: 200 }));
  vi.stubGlobal("fetch", fetchMock);
  await updateGallery(gallery.id, {
    name: gallery.name,
    description: gallery.description,
    visibility: gallery.visibility,
    wallRenderStrategy: "grid",
  });
  expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual({
    name: gallery.name,
    description: gallery.description,
    visibility: gallery.visibility,
    wallRenderStrategy: "grid",
  });
});

test("sharing by username uses the resolver route", async () => {
  const fetchMock = vi
    .fn()
    .mockResolvedValue(new Response(JSON.stringify({ userId: "1002" }), { status: 200 }));
  vi.stubGlobal("fetch", fetchMock);
  await shareGalleryByUsername(gallery.id, "oleksiybondar");
  expect({ path: fetchMock.mock.calls[0][0], body: fetchMock.mock.calls[0][1].body }).toEqual({
    path: `/api/galleries/${gallery.id}/shares`,
    body: JSON.stringify({ username: "oleksiybondar" }),
  });
});

test("adding a folder uses the collection directory route", async () => {
  const fetchMock = vi.fn().mockResolvedValue(new Response("{}", { status: 200 }));
  vi.stubGlobal("fetch", fetchMock);
  await addCollectionDirectory(gallery.id, "/NASPool/Photos");
  expect({ path: fetchMock.mock.calls[0][0], method: fetchMock.mock.calls[0][1].method }).toEqual({
    path: `/api/collections/${gallery.id}/directories`,
    method: "POST",
  });
});

test("removing an album folder uses the collection directory route", async () => {
  const fetchMock = vi.fn().mockResolvedValue(new Response("{}", { status: 200 }));
  vi.stubGlobal("fetch", fetchMock);
  await removeCollectionDirectory(gallery.id, "/NASPool/Photos");
  expect({ path: fetchMock.mock.calls[0][0], method: fetchMock.mock.calls[0][1].method }).toEqual({
    path: `/api/collections/${gallery.id}/directories`,
    method: "DELETE",
  });
});

test("removing an album folder rejects a relative path before transport", async () => {
  const fetchMock = vi.fn();
  vi.stubGlobal("fetch", fetchMock);
  await expect(removeCollectionDirectory(gallery.id, "Photos")).rejects.toThrow();
});

test("saving a generated cover uploads the image bytes to the gallery", async () => {
  const fetchMock = vi
    .fn()
    .mockResolvedValue(new Response(JSON.stringify(gallery), { status: 200 }));
  vi.stubGlobal("fetch", fetchMock);
  const image = new Blob(["png"], { type: "image/png" });
  await saveGalleryThumbnail(gallery.id, image);
  expect({
    path: fetchMock.mock.calls[0][0],
    method: fetchMock.mock.calls[0][1].method,
    contentType: fetchMock.mock.calls[0][1].headers["Content-Type"],
    body: fetchMock.mock.calls[0][1].body,
  }).toEqual({
    path: `/api/galleries/${gallery.id}/thumbnail/image`,
    method: "PUT",
    contentType: "image/png",
    body: image,
  });
});

test("selecting a NAS image persists its path through gallery metadata", async () => {
  const fetchMock = vi.fn().mockResolvedValue(
    new Response(JSON.stringify({ ...gallery, thumbnailPath: "/home/Photos/cover.jpg" }), {
      status: 200,
    }),
  );
  vi.stubGlobal("fetch", fetchMock);
  await chooseGalleryThumbnailPath(gallery.id, "/home/Photos/cover.jpg");
  expect(fetchMock.mock.calls[0][1].body).toBe(
    JSON.stringify({ thumbnailPath: "/home/Photos/cover.jpg" }),
  );
});
