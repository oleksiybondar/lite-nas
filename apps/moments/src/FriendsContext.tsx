import type { ReactElement, ReactNode } from "react";
import { createContext, useCallback, useContext, useRef, useState } from "react";
import type { z } from "zod";
import {
  listFriends as listAllFriends,
  listFriendsPage,
  removeFriend,
  saveFriend,
} from "./moments-data";
import type { friendInputSchema, friendPageSchema } from "./schemas";

type FriendPage = z.infer<typeof friendPageSchema>;
type FriendInput = z.input<typeof friendInputSchema>;

/** Shared contact data and validated operations used by Friends and gallery sharing. */
interface FriendsContextValue {
  pageData: FriendPage | null;
  loading: boolean;
  loadFailed: boolean;
  loadFriends: (page: number, size: number, query: string) => Promise<FriendPage | null>;
  loadAllFriends: () => Promise<FriendPage["items"]>;
  save: (friend: FriendInput) => Promise<void>;
  remove: (username: string) => Promise<number>;
}

const FriendsContext = createContext<FriendsContextValue | null>(null);

/** Keeps the active paginated contact result available across authenticated Moments screens. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: provider callbacks share one Friends state lifecycle.
export const FriendsProvider = ({ children }: { children: ReactNode }): ReactElement => {
  const [pageData, setPageData] = useState<FriendPage | null>(null);
  const [loading, setLoading] = useState(false);
  const [loadFailed, setLoadFailed] = useState(false);
  const requestId = useRef(0);
  const activeQuery = useRef({ page: 1, size: 20, query: "" });
  const currentItems = useRef<FriendPage["items"]>([]);

  const loadFriends = useCallback(
    // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: request staleness and loading/error transitions must stay coordinated.
    async (page: number, size: number, query: string): Promise<FriendPage | null> => {
      activeQuery.current = { page, size, query };
      const id = ++requestId.current;
      setLoading(true);
      setLoadFailed(false);
      try {
        const result = await listFriendsPage(page, size, query);
        if (id !== requestId.current) return null;
        setPageData(result);
        currentItems.current = result.items;
        return result;
      } catch {
        if (id === requestId.current) setLoadFailed(true);
        return null;
      } finally {
        if (id === requestId.current) setLoading(false);
      }
    },
    [],
  );
  const loadAllFriends = useCallback(() => listAllFriends(), []);

  const save = useCallback(
    async (friend: FriendInput): Promise<void> => {
      await saveFriend(friend);
      const query = activeQuery.current;
      await loadFriends(query.page, query.size, query.query);
    },
    [loadFriends],
  );

  const remove = useCallback(
    async (username: string): Promise<number> => {
      await removeFriend(username);
      const query = activeQuery.current;
      const page =
        query.page > 1 && currentItems.current.length === 1 ? query.page - 1 : query.page;
      await loadFriends(page, query.size, query.query);
      return page;
    },
    [loadFriends],
  );

  return (
    <FriendsContext.Provider
      value={{ pageData, loading, loadFailed, loadFriends, loadAllFriends, save, remove }}
    >
      {children}
    </FriendsContext.Provider>
  );
};

/** Returns shared contacts and operations; requires the authenticated FriendsProvider. */
export const useFriends = (): FriendsContextValue => {
  const value = useContext(FriendsContext);
  if (!value) throw new Error("FriendsProvider is required");
  return value;
};
