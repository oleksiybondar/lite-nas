import { LanguageProvider } from "@lite-nas/language";
import { render } from "@testing-library/react";
import type { PropsWithChildren, ReactElement } from "react";

/** Supplies the default browser language to isolated admin component tests. */
export const TestLanguageBoundary = ({ children }: PropsWithChildren): ReactElement => (
  <LanguageProvider>{children}</LanguageProvider>
);

/** Renders an isolated component with the English browser fallback. */
export const renderWithLanguage = (component: ReactElement): ReturnType<typeof render> =>
  render(component, { wrapper: TestLanguageBoundary });
