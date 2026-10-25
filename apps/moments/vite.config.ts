import { resolve } from "node:path";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

const fileServiceOrigin = process.env.LITE_NAS_FILE_SERVICE_ORIGIN || "http://127.0.0.1:8082";
const streamServiceOrigin = process.env.LITE_NAS_STREAM_SERVICE_ORIGIN || "http://127.0.0.1:8083";

/** Builds relative assets so the same bundle serves standalone and under /apps/moments/. */
export default defineConfig({
  base: "./",
  build: {
    emptyOutDir: true,
    outDir: process.env.LITE_NAS_MOMENTS_OUT_DIR || "../../.build/moments",
  },
  plugins: [react()],
  test: {
    coverage: {
      allowExternal: true,
      include: [
        resolve("src/**/*.{ts,tsx}"),
        resolve("../../shared/ts/file-browser/FileSelectionPage.tsx"),
      ],
      thresholds: { branches: 75, functions: 75, lines: 75, statements: 75 },
    },
  },
  server: {
    proxy: {
      "/api": {
        target: process.env.LITE_NAS_MOMENTS_API_ORIGIN || "http://127.0.0.1:9091",
        changeOrigin: false,
      },
      "/f/": { target: fileServiceOrigin, changeOrigin: false, rewrite: (path) => path.slice(2) },
      "/d/": {
        target: fileServiceOrigin,
        changeOrigin: false,
        rewrite: (path) => `/download${path.slice(2)}`,
      },
      "/ls/": { target: fileServiceOrigin, changeOrigin: false },
      "/s/": {
        target: streamServiceOrigin,
        changeOrigin: false,
        rewrite: (path) => path.slice(2),
      },
    },
  },
});
