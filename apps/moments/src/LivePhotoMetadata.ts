import type { GalleryItem } from "./gallery-model";

/** Shortlists unambiguous HEIC/HEIF and MOV companions in the same folder. */
// biome-ignore lint/complexity/noExcessiveCognitiveComplexity: filename matching and ambiguity detection form one collection pass.
const candidatePairs = (items: readonly GalleryItem[]): Array<[GalleryItem, GalleryItem]> => {
  const byStem = new Map<
    string,
    { still?: GalleryItem; motion?: GalleryItem; ambiguous?: boolean }
  >();
  for (const item of items) {
    const match = item.url.match(/^(.*\/)([^/]+)\.(heic|heif|mov)$/i);
    if (!match) continue;
    const key = `${match[1]}${match[2].toLowerCase()}`;
    const entry = byStem.get(key) ?? {};
    const field = /\.mov$/i.test(item.url) ? "motion" : "still";
    if (entry[field]) entry.ambiguous = true;
    else entry[field] = item;
    byStem.set(key, entry);
  }
  return [...byStem.values()]
    .filter((entry) => !entry.ambiguous && entry.still && entry.motion)
    .map((entry) => [entry.still as GalleryItem, entry.motion as GalleryItem]);
};

/** Keeps standalone media and replaces each filename-matched pair with its still. */
export const composeLivePhotoItems = (items: readonly GalleryItem[]): GalleryItem[] => {
  const supportedItems = items.filter((item) => item.kind !== "unsupported");
  const paired = new Map(
    candidatePairs(supportedItems).map(([still, motion]) => [still.id, motion.url]),
  );
  const motionUrls = new Set(paired.values());
  return supportedItems.flatMap((item) => {
    if (motionUrls.has(item.url)) return [];
    const motionUrl = paired.get(item.id);
    return [motionUrl ? { ...item, motionUrl } : item];
  });
};
