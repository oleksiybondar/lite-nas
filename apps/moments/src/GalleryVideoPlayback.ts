/** Outcome of attempting browser-controlled video autoplay. */
export type VideoAutoplayResult = "started" | "blocked" | "disabled" | "interrupted" | "failed";

/** Maps browser playback rejection names to stable viewer outcomes. */
const rejectionResult = (error: unknown): VideoAutoplayResult => {
  const name = error instanceof Error ? error.name : "";
  if (name === "NotAllowedError") return "blocked";
  if (name === "AbortError") return "interrupted";
  return "failed";
};

/** Attempts playback without treating an autoplay-policy rejection as a media error. */
export const attemptVideoAutoplay = async (
  video: Pick<HTMLVideoElement, "play">,
  enabled: boolean,
): Promise<VideoAutoplayResult> => {
  if (!enabled) return "disabled";
  try {
    await video.play();
    return "started";
  } catch (error) {
    return rejectionResult(error);
  }
};
