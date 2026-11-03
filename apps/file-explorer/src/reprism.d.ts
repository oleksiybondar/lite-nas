declare module "reprism" {
  /** One RePrism token or unclassified source segment returned by tokenize. */
  export type TokenizedSource =
    | string
    | {
        type: string;
        content: TokenizedSource | TokenizedSource[];
        alias?: string | string[];
      };

  /** Grammar module consumed by RePrism's language registry. */
  export interface LanguageDefinition {
    language: string;
    init(prism: unknown): void;
  }

  /** Tokenizer and grammar registry for constructing safe React token spans. */
  const Prism: {
    languages: Record<string, unknown>;
    tokenize(source: string, grammar: unknown): TokenizedSource[];
  };
  export default Prism;

  /** Registers one or more language grammars for later highlighting. */
  export function loadLanguages(languages: LanguageDefinition[]): void;
}

declare module "reprism/languages/*" {
  import type { LanguageDefinition } from "reprism";

  const language: LanguageDefinition;
  export default language;
}
