import { z } from "zod";

const locationSchema = z.object({
  latitude: z.number().finite().min(-90).max(90),
  longitude: z.number().finite().min(-180).max(180),
});

/** Builds an explicit Google Maps search link from a validated GPS pair. */
export const googleMapsUrl = (location: { latitude: number; longitude: number }): string | null => {
  const result = locationSchema.safeParse(location);
  if (!result.success) return null;
  const query = `${result.data.latitude.toFixed(6)},${result.data.longitude.toFixed(6)}`;
  return `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(query)}`;
};
