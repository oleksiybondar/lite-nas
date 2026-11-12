import type { ReactElement, ReactNode } from "react";
import { createContext, useContext } from "react";
import { galleryFileUrlSchema, galleryIsHeic } from "./gallery-model";

/** URL contract that lets HTML5 media elements request authorized host files. */
export type StreamProtocol = {
  original: (fileUrl: string) => string | null;
  transcode: (fileUrl: string, start: number) => string | null;
  image: (fileUrl: string) => string | null;
};

/** Validates File Service URLs before deriving Stream Service media routes. */
export const streamProtocol: StreamProtocol = {
  original: (fileUrl) => {
    const result = galleryFileUrlSchema.safeParse(fileUrl);
    return result.success ? `/s/${result.data.slice(3)}` : null;
  },
  transcode: (fileUrl, start) => {
    const result = galleryFileUrlSchema.safeParse(fileUrl);
    if (!result.success || !Number.isSafeInteger(start) || start < 0 || start > 86400) return null;
    return `/s/transcode/${result.data.slice(3)}?start=${start}`;
  },
  image: (fileUrl) => {
    const result = galleryFileUrlSchema.safeParse(fileUrl);
    if (!result.success) return null;
    return galleryIsHeic(result.data) ? `/s/preview/${result.data.slice(3)}` : result.data;
  },
};

const StreamContext = createContext<StreamProtocol | null>(null);

/** Makes the media URL contract available throughout the Moments app. */
export const StreamProvider = ({ children }: { children: ReactNode }): ReactElement => (
  <StreamContext.Provider value={streamProtocol}>{children}</StreamContext.Provider>
);

/** Returns the validated media URL contract for HTML5 elements. */
export const useStream = (): StreamProtocol => {
  const protocol = useContext(StreamContext);
  if (!protocol) throw new Error("StreamProvider is required");
  return protocol;
};
