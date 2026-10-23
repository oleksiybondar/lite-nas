import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

/** Builds assets for the Nginx mount owned by the File Explorer package. */
export default defineConfig({
  base: "/apps/file-explorer/",
  build: {
    emptyOutDir: true,
    outDir: process.env.LITE_NAS_FILE_EXPLORER_OUT_DIR || "../../.build/file-explorer",
  },
  plugins: [react()],
});
