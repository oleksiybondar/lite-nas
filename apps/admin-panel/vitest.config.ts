import path from "node:path";

import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@app": path.resolve(__dirname, "src/app"),
      "@assets": path.resolve(__dirname, "src/assets"),
      "@components": path.resolve(__dirname, "src/components"),
      "@configs": path.resolve(__dirname, "src/configs"),
      "@contexts": path.resolve(__dirname, "src/contexts"),
      "@domain": path.resolve(__dirname, "src/domain"),
      "@helpers": path.resolve(__dirname, "src/helpers"),
      "@hooks": path.resolve(__dirname, "src/hooks"),
      "@models": path.resolve(__dirname, "src/models"),
      "@pages": path.resolve(__dirname, "src/pages"),
      "@providers": path.resolve(__dirname, "src/providers"),
      "@routes": path.resolve(__dirname, "src/routes"),
      "@tests": path.resolve(__dirname, "tests"),
      "@theme": path.resolve(__dirname, "src/theme"),
    },
  },
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["tests/setup/vitest.setup.ts"],
  },
});
