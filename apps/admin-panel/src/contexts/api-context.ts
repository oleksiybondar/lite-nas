import { createContext } from "react";
import type { ApiContextValue } from "../dto/api/api";

/**
 * Context for the app API client.
 *
 * Consumers should use `useApi` instead of reading this context directly. The
 * provider supplies method-specific fetch wrappers and centralizes app-level
 * unauthorized handling so feature components do not duplicate refresh and
 * login navigation logic.
 */
export const ApiContext = createContext<ApiContextValue | undefined>(undefined);
