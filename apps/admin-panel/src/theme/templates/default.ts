import type { ThemeOptions } from "@mui/material/styles";
import type { ThemeTemplate } from "@theme/types";

const fontFamily = [
  "Inter",
  "-apple-system",
  "BlinkMacSystemFont",
  '"Segoe UI"',
  "sans-serif",
].join(",");

const commonTemplateOptions = {
  shape: {
    borderRadius: 8,
  },
  typography: {
    fontFamily,
    h1: {
      fontSize: "2rem",
      fontWeight: 700,
      lineHeight: 1.15,
    },
    h2: {
      fontSize: "1.5rem",
      fontWeight: 700,
      lineHeight: 1.2,
    },
    overline: {
      fontWeight: 700,
      letterSpacing: "0.08em",
    },
  },
} satisfies Pick<ThemeOptions, "shape" | "typography">;

export const defaultTheme: ThemeTemplate = {
  dark: {
    ...commonTemplateOptions,
    palette: {
      mode: "dark",
      background: {
        default: "#111827",
        paper: "#1f2937",
      },
      divider: "rgba(148, 163, 184, 0.24)",
      primary: {
        main: "#90caf9",
      },
      secondary: {
        main: "#f48fb1",
      },
      text: {
        primary: "#f9fafb",
        secondary: "#cbd5e1",
      },
    },
  },
  light: {
    ...commonTemplateOptions,
    palette: {
      mode: "light",
      background: {
        default: "#f7f8fc",
        paper: "#ffffff",
      },
      primary: {
        main: "#1d4ed8",
      },
      secondary: {
        main: "#7c3aed",
      },
      text: {
        primary: "#111827",
        secondary: "#475569",
      },
    },
  },
};
