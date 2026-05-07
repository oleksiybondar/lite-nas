import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

import { createViteAliases } from "./vite.aliases";

/**
 * Default local gateway origin used by the Vite development server.
 */
const defaultGatewayOrigin = "http://127.0.0.1:9090";

export default defineConfig(() => {
  const outDir = process.env.LITE_NAS_ADMIN_PANEL_OUT_DIR || "../../.build/admin-panel";
  const gatewayOrigin = process.env.LITE_NAS_WEB_GATEWAY_ORIGIN || defaultGatewayOrigin;

  return {
    build: {
      chunkSizeWarningLimit: 700,
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
          target: gatewayOrigin,
        },
      },
    },
  };
});
