import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

import { createViteAliases } from "./vite.aliases";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: createViteAliases(__dirname),
  },
  test: {
    coverage: {
      include: ["src/**/*.{ts,tsx}"],
      thresholds: {
        branches: 75,
        functions: 75,
        lines: 75,
        statements: 75,
      },
    },
    environment: "jsdom",
    globals: true,
    setupFiles: ["tests/setup/vitest.setup.ts"],
  },
});
