declare module "reprism" {
  /** Registers modular Prism grammar definitions before highlighting code. */
  export function loadLanguages(
    ...languages: Array<{ language: string; init: (prism: unknown) => void }>
  ): void;
  /** Options accepted by reprism's markup-generating highlighter. */
  export interface HighlightOptions {
    component?: string | false;
  }
  /** Returns syntax-highlighted markup for a source string. */
  export function highlight(code: string, language: string, options?: HighlightOptions): string;
}

declare module "reprism/languages/ini" {
  const language: { language: string; init: (prism: unknown) => void };
  export default language;
}

declare module "reprism/languages/json" {
  const language: { language: string; init: (prism: unknown) => void };
  export default language;
}

declare module "reprism/languages/yaml" {
  const language: { language: string; init: (prism: unknown) => void };
  export default language;
}
