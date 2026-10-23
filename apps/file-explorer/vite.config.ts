import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

/** Default local File Service origin for requests made through the Vite server. */
const defaultFileServiceOrigin = "http://127.0.0.1:8082";

/** Builds assets for the Nginx mount owned by the File Explorer package. */
export default defineConfig({
  base: "/apps/file-explorer/",
  build: {
    emptyOutDir: true,
    outDir: process.env.LITE_NAS_FILE_EXPLORER_OUT_DIR || "../../.build/file-explorer",
  },
  plugins: [react()],
  server: {
    proxy: {
      "/f/": {
        target: process.env.LITE_NAS_FILE_SERVICE_ORIGIN || defaultFileServiceOrigin,
        rewrite: (path) => path.slice(2),
      },
    },
  },
});
