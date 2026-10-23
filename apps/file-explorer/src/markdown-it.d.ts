declare module "markdown-it" {
  /** Markdown renderer configured by File Explorer for safe preview output. */
  class MarkdownIt {
    constructor(options?: {
      html?: boolean;
      linkify?: boolean;
      typographer?: boolean;
    });

    /** Renders Markdown source as HTML. */
    render(source: string): string;
  }

  export default MarkdownIt;
}
