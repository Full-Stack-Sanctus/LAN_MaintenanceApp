import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue"; // Ensure you have your vue plugin imported

const host = process.env.TAURI_DEV_HOST;

export default defineConfig(async () => ({
  plugins: [vue()],

  // 1. Set base to './' so assets use relative paths
  base: './', 

  // 2. Ensure Vite outputs to the directory Tauri is looking at
  build: {
    outDir: '../dist',
    emptyOutDir: true,
    // Tauri supports modern browsers, so we can target high-end JS
    target: process.env.TAURI_ENV_PLATFORM == 'windows' ? 'chrome105' : 'safari13',
    // Generate source maps only if needed for debugging
    sourcemap: !!process.env.TAURI_ENV_DEBUG,
  },

  clearScreen: false,
  server: {
    port: 1420,
    strictPort: true,
    host: host || false,
    hmr: host
      ? {
          protocol: "ws",
          host,
          port: 1421,
        }
      : undefined,
    watch: {
      ignored: ["**/src-tauri/**"],
    },
  },
}));