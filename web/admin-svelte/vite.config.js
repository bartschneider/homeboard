import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";

export default defineConfig({
  base: "/static/admin/",
  plugins: [svelte()],
  build: {
    outDir: "../../static/admin",
    emptyOutDir: true
  },
  server: {
    port: 5173,
    proxy: {
      "/api": "http://localhost:8081",
      "/ws": {
        target: "http://localhost:8081",
        ws: true
      }
    }
  }
});
