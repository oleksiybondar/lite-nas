import { AppProviders } from "@providers/AppProviders";
import { AppRouter } from "@routes/AppRouter";
import type { ReactElement } from "react";

export const App = (): ReactElement => {
  return (
    <div data-testid="app-root">
      <AppProviders>
        <AppRouter />
      </AppProviders>
    </div>
  );
};
