import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

import { createViteAliases } from "./vite.aliases";

export default defineConfig(() => {
  const outDir = process.env.LITE_NAS_ADMIN_PANEL_OUT_DIR || "../../.build/admin-panel";

  return {
    build: {
      emptyOutDir: true,
      outDir,
      rollupOptions: {
        output: {
          assetFileNames: "assets/[name][extname]",
          chunkFileNames: "assets/[name].js",
          entryFileNames: "assets/index.js",
        },
      },
    },
    plugins: [react()],
    resolve: {
      alias: createViteAliases(__dirname),
    },
  };
});
