import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

import { createViteAliases } from "./vite.aliases";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: createViteAliases(__dirname),
  },
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["tests/setup/vitest.setup.ts"],
  },
});
