import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

/** Builds relative assets so the same bundle serves standalone and under /apps/moments/. */
export default defineConfig({
  base: "./",
  build: {
    emptyOutDir: true,
    outDir: process.env.LITE_NAS_MOMENTS_OUT_DIR || "../../.build/moments",
  },
  plugins: [react()],
  server: {
    proxy: {
      "/api": process.env.LITE_NAS_MOMENTS_API_ORIGIN || "http://127.0.0.1:9091",
      "/f/": {
        target: process.env.LITE_NAS_FILE_SERVICE_ORIGIN || "http://127.0.0.1:8082",
        rewrite: (path) => path.slice(2),
      },
      "/d/": {
        target: process.env.LITE_NAS_FILE_SERVICE_ORIGIN || "http://127.0.0.1:8082",
        rewrite: (path) => `/download${path.slice(2)}`,
      },
      "/ls/": process.env.LITE_NAS_FILE_SERVICE_ORIGIN || "http://127.0.0.1:8082",
    },
  },
});
