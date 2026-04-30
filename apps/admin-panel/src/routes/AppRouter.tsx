import { router } from "@routes/router";
import type { ReactElement } from "react";
import { RouterProvider } from "react-router-dom";

export const AppRouter = (): ReactElement => {
  return <RouterProvider router={router} />;
};
