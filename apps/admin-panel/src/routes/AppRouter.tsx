import { router } from "@routes/router";
import type { ReactElement } from "react";
import { RouterProvider } from "react-router-dom";

/**
 * App-level data router configured with React Router v7 transition behavior.
 */
export const AppRouter = (): ReactElement => {
  return <RouterProvider future={{ v7_startTransition: true }} router={router} />;
};
