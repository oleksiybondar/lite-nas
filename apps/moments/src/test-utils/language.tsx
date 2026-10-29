import { LanguageProvider } from "@lite-nas/language";
import { render as testingLibraryRender } from "@testing-library/react";
import type { ReactElement, ReactNode } from "react";

/** Renders an isolated Moments component with the normal English fallback language. */
export const renderWithLanguage = (component: ReactElement) =>
  testingLibraryRender(component, { wrapper: LanguageBoundary });

/** Keeps the language context attached to every rerender of a test tree. */
const LanguageBoundary = ({ children }: { children: ReactNode }): ReactElement => (
  <LanguageProvider>{children}</LanguageProvider>
);
