/** Limits original-image downloads shared by the wall and active viewer. */
export class GalleryImageQueue {
  /** Bounds simultaneous original downloads for this gallery instance. */
  constructor(readonly maxConcurrent = 10) {}
  private readonly listeners = new Set<() => void>();
  private readonly ready = new Map<string, number>();
  private readonly failed = new Set<string>();
  private readonly active = new Map<string, number>();
  private desired: string[] = [];
  private version = 0;
  private nextAttempt = 0;

  /** Returns a stable primitive snapshot for React subscriptions. */
  getSnapshot = (): number => this.version;

  /** Subscribes a view to image readiness changes. */
  subscribe = (listener: () => void): (() => void) => {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  };

  /** Reports whether an image element may start or has finished its request. */
  status = (url: string): "ready" | "error" | "loading" | "queued" => {
    if (this.ready.has(url)) return "ready";
    if (this.failed.has(url)) return "error";
    if (this.active.has(url)) return "loading";
    return "queued";
  };

  /** Identifies the current element request so stale load events can be ignored. */
  attempt = (url: string): number | undefined => this.active.get(url) ?? this.ready.get(url);

  /** Replaces pending work with currently needed URLs, ordered by priority. */
  want = (urls: readonly string[]): void => {
    const next = [...new Set(urls)];
    if (
      next.length === this.desired.length &&
      next.every((url, index) => url === this.desired[index])
    )
      return;
    this.desired = next;
    for (const url of this.active.keys()) if (!next.includes(url)) this.active.delete(url);
    this.pump();
  };

  /** Drops completed state once no wall or viewer element retains a URL. */
  retain = (urls: readonly string[]): void => {
    const retained = new Set(urls);
    const readyChanged = this.removeUnretained(this.ready.keys(), retained, (url) =>
      this.ready.delete(url),
    );
    const failedChanged = this.removeUnretained(this.failed.values(), retained, (url) =>
      this.failed.delete(url),
    );
    if (readyChanged || failedChanged) this.changed();
  };

  /** Settles a request made by an image element and grants the next slot. */
  complete = (url: string, success: boolean, attempt = this.active.get(url)): void => {
    if (attempt === undefined || this.active.get(url) !== attempt) return;
    this.active.delete(url);
    if (success) this.ready.set(url, attempt);
    else this.failed.add(url);
    this.pump();
  };

  /** Ignores stale errors from canceled elements and records current failures. */
  fail = (url: string, attempt = this.attempt(url)): void => {
    if (attempt === undefined) return;
    if (this.active.get(url) === attempt) {
      this.complete(url, false, attempt);
      return;
    }
    if (this.ready.get(url) !== attempt || !this.desired.includes(url)) return;
    this.ready.delete(url);
    this.failed.add(url);
    this.changed();
  };

  /** Retries a failed original only after an explicit viewer action. */
  retry = (url: string): void => {
    if (!this.failed.delete(url)) return;
    this.pump();
  };

  /** Releases pending work when the gallery leaves the page. */
  dispose = (): void => {
    this.listeners.clear();
    this.want([]);
  };

  /** Publishes a new request-state snapshot to mounted gallery views. */
  private changed(): void {
    this.version += 1;
    for (const listener of this.listeners) listener();
  }

  /** Removes values excluded by the current wall and viewer window. */
  private removeUnretained(
    values: IterableIterator<string>,
    retained: ReadonlySet<string>,
    remove: (url: string) => boolean,
  ): boolean {
    let changed = false;
    for (const url of values) if (!retained.has(url)) changed = remove(url) || changed;
    return changed;
  }

  /** Grants at most the configured number of original download slots. */
  private pump(): void {
    for (const url of this.desired) {
      if (this.active.size >= this.maxConcurrent) break;
      if (this.isSettledOrActive(url)) continue;
      this.active.set(url, ++this.nextAttempt);
    }
    this.changed();
  }

  /** Identifies requests that already own a slot or have a terminal status. */
  private isSettledOrActive(url: string): boolean {
    return this.ready.has(url) || this.failed.has(url) || this.active.has(url);
  }
}
