import type { PropsWithChildren, ReactElement } from "react";
import { MemoryRouter } from "react-router-dom";

/**
 * Shared React Router future flags used by unit tests to silence v7 transition warnings.
 */
const memoryRouterFuture = {
  v7_relativeSplatPath: true,
  v7_startTransition: true,
} as const;

type TestMemoryRouterProps = PropsWithChildren<{
  /**
   * Optional initial entries forwarded to React Router's memory router.
   */
  initialEntries?: string[];
}>;

/**
 * Test-only memory router configured with the React Router v7 future flags.
 */
export const TestMemoryRouter = ({
  children,
  initialEntries,
}: TestMemoryRouterProps): ReactElement => {
  return (
    <MemoryRouter
      future={memoryRouterFuture}
      {...(initialEntries !== undefined ? { initialEntries } : {})}
    >
      {children}
    </MemoryRouter>
  );
};
