import react from "@vitejs/plugin-react";
import { defineConfig, loadEnv } from "vite";

import { createViteAliases } from "./vite.aliases";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const backendTarget = env.API_URL || "http://127.0.0.1:8080";
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
    server: {
      proxy: {
        "/api": {
          changeOrigin: true,
          target: backendTarget,
        },
      },
    },
  };
});
