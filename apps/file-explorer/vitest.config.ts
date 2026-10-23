import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

/** Runs File Explorer tests and enforces the repository's TypeScript coverage floor. */
export default defineConfig({
  plugins: [react()],
  test: {
    environment: "jsdom",
    setupFiles: ["tests/setup.ts"],
    coverage: {
      include: ["src/**/*.{ts,tsx}"],
      thresholds: { branches: 75, functions: 75, lines: 75, statements: 75 },
    },
  },
});
