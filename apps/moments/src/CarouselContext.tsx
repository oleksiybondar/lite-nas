// biome-ignore-all lint/complexity/noExcessiveLinesPerFunction: state transitions and their timer stay together in this provider.
import type { Dispatch, ReactElement, ReactNode, SetStateAction } from "react";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  useSyncExternalStore,
} from "react";
import { useGallery } from "./GalleryContext";
import type { GalleryItem } from "./gallery-model";

/** Viewer-only selection, playback, and slideshow controls. */
export type CarouselState = {
  selected: number | null;
  current: GalleryItem | null;
  playing: boolean;
  countdown: number | null;
  videoAutoplay: boolean;
  setVideoAutoplay: (enabled: boolean) => void;
  setPlaying: (playing: boolean) => void;
  delay: 1000 | 3000 | 5000;
  setDelay: (delay: 1000 | 3000 | 5000) => void;
  videoError: boolean;
  setVideoError: (error: boolean) => void;
  open: (index: number) => void;
  close: () => void;
  navigate: (step: number) => void;
};
const CarouselContext = createContext<CarouselState | null>(null);

/** Reports whether the selected still can advance the timed slideshow. */
const canAdvanceSlideshow = (
  playing: boolean,
  selected: number | null,
  current: GalleryItem | null,
  status: string | null,
): boolean =>
  playing &&
  selected !== null &&
  current !== null &&
  current.kind !== "video" &&
  (current.kind !== "image" || status === "ready" || status === "error");

/** Owns the slideshow countdown and advances only after an image is settled. */
const useSlideshowTimer = ({
  playing,
  selected,
  current,
  status,
  delay,
  length,
  setPlaying,
  setCountdown,
  setSelected,
  setVideoError,
}: {
  playing: boolean;
  selected: number | null;
  current: GalleryItem | null;
  status: string | null;
  delay: number;
  length: number;
  setPlaying: (playing: boolean) => void;
  setCountdown: (countdown: number | null) => void;
  setSelected: Dispatch<SetStateAction<number | null>>;
  setVideoError: (error: boolean) => void;
}): void => {
  useEffect(() => {
    if (!canAdvanceSlideshow(playing, selected, current, status)) {
      setCountdown(null);
      return;
    }
    if (selected === length - 1) {
      setCountdown(null);
      setPlaying(false);
      return;
    }
    let remaining = delay / 1000;
    setCountdown(remaining);
    const timer = window.setInterval(() => {
      remaining -= 1;
      if (remaining > 0) {
        setCountdown(remaining);
        return;
      }
      window.clearInterval(timer);
      setCountdown(null);
      setSelected((index) => (index === null ? null : Math.min(length - 1, index + 1)));
      setVideoError(false);
    }, 1000);
    return () => window.clearInterval(timer);
  }, [
    current,
    delay,
    length,
    playing,
    selected,
    setCountdown,
    setPlaying,
    setSelected,
    setVideoError,
    status,
  ]);
};

/** Owns modal selection and slideshow timing independently of the wall. */
export const CarouselProvider = ({ children }: { children: ReactNode }): ReactElement => {
  const { items, queue } = useGallery();
  const [selected, setSelected] = useState<number | null>(null);
  const [playing, setPlaying] = useState(false);
  const [countdown, setCountdown] = useState<number | null>(null);
  const [videoAutoplay, setVideoAutoplay] = useState(true);
  const [delay, setDelay] = useState<1000 | 3000 | 5000>(3000);
  const [videoError, setVideoError] = useState(false);
  const subscribePlayback = useCallback(
    (listener: () => void) => {
      if (!playing || selected === null) return () => {};
      return queue.subscribe(listener);
    },
    [playing, selected, queue],
  );
  useSyncExternalStore(subscribePlayback, queue.getSnapshot);
  const current = selected === null ? null : (items[selected] ?? null);
  const status = current?.kind === "image" ? queue.status(current.url) : null;
  /** Opens one media item without changing wall geometry. */
  const open = (index: number): void => {
    setSelected(index);
    setCountdown(null);
    setVideoError(false);
  };
  /** Closes the modal and stops timed navigation. */
  const close = (): void => {
    setSelected(null);
    setPlaying(false);
    setCountdown(null);
  };
  /** Moves by one item without wrapping past the collection ends. */
  const navigate = (step: number): void => {
    setSelected((index) =>
      index === null ? null : Math.max(0, Math.min(items.length - 1, index + step)),
    );
    setCountdown(null);
    setVideoError(false);
  };
  useEffect(() => {
    if (!videoError || !playing || selected === null || current?.kind !== "video") return;
    if (selected === items.length - 1) {
      setPlaying(false);
      return;
    }
    setSelected(selected + 1);
    setVideoError(false);
  }, [current, items.length, playing, selected, videoError]);
  useSlideshowTimer({
    playing,
    selected,
    current,
    status,
    delay,
    length: items.length,
    setPlaying,
    setCountdown,
    setSelected,
    setVideoError,
  });
  const value = {
    selected,
    current,
    playing,
    countdown,
    videoAutoplay,
    setVideoAutoplay,
    setPlaying,
    delay,
    setDelay,
    videoError,
    setVideoError,
    open,
    close,
    navigate,
  };
  return <CarouselContext.Provider value={value}>{children}</CarouselContext.Provider>;
};

/** Reads viewer state without sharing its timer with wall consumers. */
export const useCarousel = (): CarouselState => {
  const value = useContext(CarouselContext);
  if (!value) throw new Error("CarouselProvider is required");
  return value;
};
