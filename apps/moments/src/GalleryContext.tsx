import type { ReactElement, ReactNode } from "react";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useReducer,
  useRef,
} from "react";
import { GalleryImageQueue } from "./GalleryImageQueue";
import { GalleryMetadataReader } from "./GalleryMetadataReader";
import type { GalleryItem } from "./gallery-model";
import { composeLivePhotoItems } from "./LivePhotoMetadata";
import { defaultImageUiCooldownMs, type GalleryRecord } from "./schemas";

/** Stable collection metadata and one image request coordinator shared by both views. */
export type GalleryState = {
  items: readonly GalleryItem[];
  queue: GalleryImageQueue;
  metadata: GalleryMetadataReader;
  collectionActions?: {
    collections: readonly GalleryRecord[];
    canRemove: (path: string) => boolean;
    add: (collectionId: string, path: string) => Promise<void>;
    remove: (path: string) => Promise<void>;
    createAndAdd: (name: string, path: string) => Promise<void>;
  };
};

/**
 * Shared wall and carousel presentation state, separate from download scheduling.
 *
 * Production diagnosis found a wall tile stuck on its skeleton with no wall `<img>` in
 * the DOM, even though Network showed a 200 response for that exact original. Opening
 * the same item in the modal could make it appear. This state keeps the two views' item
 * readiness aligned when queue scheduling changes during a request. The queue remains
 * responsible for concurrency and retries; this state only controls wall/strip rendering.
 */
export type GalleryMediaUiState = {
  revision: number;
  currentIndex: number | null;
  wall: { rendered: string[]; visible: string[]; direction: "up" | "down" | "idle" };
  carouselStrip: { rendered: string[]; visible: string[] };
  items: Record<string, { url: string; status: "loading" | "ready" }>;
};

/**
 * UI events that synchronize the wall and carousel without owning retries.
 *
 * Window actions carry the complete new window so adding tiles and pruning old item
 * records happen in one reducer transition. Image failures deliberately have no action:
 * the placeholder remains visible and retry policy stays with the download queue. Image
 * status publication uses the user-configured cooldown (100 ms by default); wall and
 * selection actions are dispatched immediately.
 */
export type GalleryMediaUiAction =
  | {
      type: "wall/windowChanged";
      rendered: string[];
      visible: string[];
      direction: "up" | "down" | "idle";
    }
  | {
      type: "carousel/selectionChanged";
      currentIndex: number | null;
      rendered: string[];
      visible: string[];
    }
  | { type: "image/loading"; itemId: string; url: string }
  | { type: "image/ready"; itemId: string; url: string };

/** Reducer-only batch used by the provider's private image publication cooldown. */
type GalleryMediaUiReducerAction =
  | GalleryMediaUiAction
  | {
      type: "images/batch";
      updates: Array<{ itemId: string; url: string; status: "loading" | "ready" }>;
    };

const initialMediaUi: GalleryMediaUiState = {
  revision: 0,
  currentIndex: null,
  wall: { rendered: [], visible: [], direction: "idle" },
  carouselStrip: { rendered: [], visible: [] },
  items: {},
};

/**
 * Applies one coordinated presentation update and ignores unchanged snapshots.
 *
 * A ready record is kept while either view retains its item, even if the queue later
 * drops that URL from its active request set. Revisit this invariant alongside queue
 * cancellation semantics if request prioritization or retry behavior changes.
 */
const reduceMediaUi = (
  state: GalleryMediaUiState,
  action: GalleryMediaUiReducerAction,
): GalleryMediaUiState => {
  switch (action.type) {
    case "image/loading":
    case "image/ready":
      return reduceImageState(state, action);
    case "images/batch":
      return reduceImageBatch(state, action);
    case "wall/windowChanged":
      return reduceWallWindow(state, action);
    case "carousel/selectionChanged":
      return reduceCarouselSelection(state, action);
  }
};

/** Applies one image status update unless it repeats the current ready state. */
const reduceImageState = (
  state: GalleryMediaUiState,
  action: Extract<GalleryMediaUiAction, { type: "image/loading" | "image/ready" }>,
): GalleryMediaUiState => {
  const status = action.type === "image/ready" ? "ready" : "loading";
  const previous = state.items[action.itemId];
  if (previous?.url === action.url && (previous.status === "ready" || previous.status === status))
    return state;
  return {
    ...state,
    revision: state.revision + 1,
    items: { ...state.items, [action.itemId]: { url: action.url, status } },
  };
};

/** Applies a coalesced image batch only to items retained by either view. */
const reduceImageBatch = (
  state: GalleryMediaUiState,
  action: Extract<GalleryMediaUiReducerAction, { type: "images/batch" }>,
): GalleryMediaUiState => {
  const retained = new Set([...state.wall.rendered, ...state.carouselStrip.rendered]);
  const updates = action.updates.filter(
    (update) =>
      retained.has(update.itemId) &&
      !(
        state.items[update.itemId]?.url === update.url &&
        (state.items[update.itemId]?.status === "ready" ||
          state.items[update.itemId]?.status === update.status)
      ),
  );
  if (updates.length === 0) return state;
  const items = { ...state.items };
  for (const update of updates) items[update.itemId] = update;
  return { ...state, revision: state.revision + 1, items };
};

/** Updates the wall window and prunes media records no longer rendered. */
const reduceWallWindow = (
  state: GalleryMediaUiState,
  action: Extract<GalleryMediaUiAction, { type: "wall/windowChanged" }>,
): GalleryMediaUiState => {
  const next = {
    ...state,
    wall: { rendered: action.rendered, visible: action.visible, direction: action.direction },
    items: retainMediaItems(state.items, [...action.rendered, ...state.carouselStrip.rendered]),
  };
  return sameMediaUiWindow(state, next) ? state : { ...next, revision: state.revision + 1 };
};

/** Updates the carousel selection and prunes media records no longer rendered. */
const reduceCarouselSelection = (
  state: GalleryMediaUiState,
  action: Extract<GalleryMediaUiAction, { type: "carousel/selectionChanged" }>,
): GalleryMediaUiState => {
  const next = {
    ...state,
    currentIndex: action.currentIndex,
    carouselStrip: { rendered: action.rendered, visible: action.visible },
    items: retainMediaItems(state.items, [...state.wall.rendered, ...action.rendered]),
  };
  return sameMediaUiWindow(state, next) ? state : { ...next, revision: state.revision + 1 };
};

/** Compares view membership and selection without considering item readiness. */
const sameMediaUiWindow = (state: GalleryMediaUiState, next: GalleryMediaUiState): boolean =>
  state.currentIndex === next.currentIndex &&
  state.wall.direction === next.wall.direction &&
  state.wall.rendered.join("\0") === next.wall.rendered.join("\0") &&
  state.wall.visible.join("\0") === next.wall.visible.join("\0") &&
  state.carouselStrip.rendered.join("\0") === next.carouselStrip.rendered.join("\0") &&
  state.carouselStrip.visible.join("\0") === next.carouselStrip.visible.join("\0");

/** Coalesces image readiness updates while applying wall and selection changes immediately. */
const publishMediaUiAction = (
  action: GalleryMediaUiAction,
  options: {
    cooldownMs: number;
    reduce: (action: GalleryMediaUiReducerAction) => void;
    cooldown: { current: ReturnType<typeof setTimeout> | null };
    pending: { current: Map<string, { itemId: string; url: string; status: "loading" | "ready" }> };
  },
): void => {
  if (isMediaUiWindowAction(action)) {
    options.reduce(action);
    return;
  }
  if (action.type === "image/loading" || action.type === "image/ready") {
    publishImageUiAction(action, options);
    return;
  }
};

/** Distinguishes synchronous viewport state from coalesced image state. */
const isMediaUiWindowAction = (
  action: GalleryMediaUiAction,
): action is Extract<
  GalleryMediaUiAction,
  { type: "wall/windowChanged" | "carousel/selectionChanged" }
> => action.type === "wall/windowChanged" || action.type === "carousel/selectionChanged";

/** Publishes a leading image state and coalesces later updates until the cooldown ends. */
const publishImageUiAction = (
  action: Extract<GalleryMediaUiAction, { type: "image/loading" | "image/ready" }>,
  options: {
    cooldownMs: number;
    reduce: (action: GalleryMediaUiReducerAction) => void;
    cooldown: { current: ReturnType<typeof setTimeout> | null };
    pending: { current: Map<string, { itemId: string; url: string; status: "loading" | "ready" }> };
  },
): void => {
  if (options.cooldownMs === 0) {
    flushMediaUi(options, action);
    return;
  }
  const update = {
    itemId: action.itemId,
    url: action.url,
    status: action.type === "image/ready" ? "ready" : "loading",
  } as const;
  const isPending = options.cooldown.current !== null;
  if (isPending && queuePendingImageUpdate(update, options.pending.current)) return;
  if (isPending) return;
  options.reduce(action);
  options.cooldown.current = setTimeout(() => flushMediaUi(options), options.cooldownMs);
};

/** Queues an update unless it would replace a ready image with its stale loading state. */
const queuePendingImageUpdate = (
  update: { itemId: string; url: string; status: "loading" | "ready" },
  pending: Map<string, { itemId: string; url: string; status: "loading" | "ready" }>,
): boolean => {
  const previous = pending.get(update.itemId);
  if (previous?.url === update.url && previous.status === "ready" && update.status === "loading")
    return true;
  pending.set(update.itemId, update);
  return false;
};

/** Flushes coalesced image updates, optionally after one leading image action. */
const flushMediaUi = (
  options: {
    reduce: (action: GalleryMediaUiReducerAction) => void;
    cooldown: { current: ReturnType<typeof setTimeout> | null };
    pending: { current: Map<string, { itemId: string; url: string; status: "loading" | "ready" }> };
  },
  leading?: GalleryMediaUiAction,
): void => {
  if (options.cooldown.current !== null) clearTimeout(options.cooldown.current);
  options.cooldown.current = null;
  const updates = [...options.pending.current.values()];
  options.pending.current.clear();
  if (updates.length > 0) options.reduce({ type: "images/batch", updates });
  if (leading) options.reduce(leading);
};

/** Keeps ready/loading records only while either view retains their tile. */
const retainMediaItems = (
  items: GalleryMediaUiState["items"],
  retainedIds: readonly string[],
): GalleryMediaUiState["items"] => {
  const retained = new Set(retainedIds);
  return Object.fromEntries(Object.entries(items).filter(([id]) => retained.has(id)));
};

const GalleryContext = createContext<GalleryState | null>(null);
const GalleryMediaUiContext = createContext<{
  state: GalleryMediaUiState;
  dispatch: (action: GalleryMediaUiAction) => void;
} | null>(null);

/**
 * Owns gallery data, download scheduling, and the reducer-backed wall/strip UI state.
 * The separate UI context keeps consumers that only need gallery data from rerendering
 * on every wall-window or image-readiness action.
 */
export const GalleryProvider = ({
  items,
  children,
  collectionActions,
  maxConcurrent = 10,
  imageUiCooldownMs = defaultImageUiCooldownMs,
}: {
  items: readonly GalleryItem[];
  children: ReactNode;
  collectionActions?: GalleryState["collectionActions"];
  maxConcurrent?: number;
  imageUiCooldownMs?: number;
}): ReactElement => {
  const queue = useMemo(() => new GalleryImageQueue(maxConcurrent), [maxConcurrent]);
  const metadata = useMemo(() => new GalleryMetadataReader(), []);
  const [mediaUi, reduceMediaUiAction] = useReducer(reduceMediaUi, initialMediaUi);
  const mediaUiCooldown = useRef<ReturnType<typeof setTimeout> | null>(null);
  const pendingImageUpdates = useRef(
    new Map<string, { itemId: string; url: string; status: "loading" | "ready" }>(),
  );
  /** Publishes the leading image update and one trailing batch per fixed cooldown window. */
  const dispatchMediaUi = useCallback(
    (action: GalleryMediaUiAction): void =>
      publishMediaUiAction(action, {
        cooldownMs: imageUiCooldownMs,
        reduce: reduceMediaUiAction,
        cooldown: mediaUiCooldown,
        pending: pendingImageUpdates,
      }),
    [imageUiCooldownMs],
  );
  const composedItems = useMemo(() => composeLivePhotoItems(items), [items]);
  const value = useMemo(
    () => ({ items: composedItems, queue, metadata, collectionActions }),
    [composedItems, queue, metadata, collectionActions],
  );
  const mediaUiValue = useMemo(
    () => ({ state: mediaUi, dispatch: dispatchMediaUi }),
    [mediaUi, dispatchMediaUi],
  );
  useEffect(
    () => () => {
      if (mediaUiCooldown.current !== null) clearTimeout(mediaUiCooldown.current);
      pendingImageUpdates.current.clear();
      queue.dispose();
      metadata.dispose();
    },
    [queue, metadata],
  );
  return (
    <GalleryContext.Provider value={value}>
      <GalleryMediaUiContext.Provider value={mediaUiValue}>
        {children}
      </GalleryMediaUiContext.Provider>
    </GalleryContext.Provider>
  );
};

/** Reads validated gallery metadata and its shared request coordinator. */
export const useGallery = (): GalleryState => {
  const value = useContext(GalleryContext);
  if (!value) throw new Error("GalleryProvider is required");
  return value;
};

/** Reads the shared wall and carousel media presentation reducer. */
export const useGalleryMediaUi = (): {
  state: GalleryMediaUiState;
  dispatch: (action: GalleryMediaUiAction) => void;
} => {
  const value = useContext(GalleryMediaUiContext);
  if (!value) throw new Error("GalleryProvider is required");
  return value;
};
