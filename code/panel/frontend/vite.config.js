import { defineConfig } from "vite";
import path from "path";
import inject from "@rollup/plugin-inject";

export default defineConfig({
  base: "",
  root: path.resolve(__dirname, "src"),

  resolve: {
    alias: {
      "~coreui": path.resolve(__dirname, "node_modules/@coreui/coreui-pro"),
    },
  },
  build: {
    minify: true,
    manifest: true,
    rollupOptions: {
      input: {
        endpoints: "./src/app/endpoints.js",
        main: "./src/app/main.js",
        deniedTokens: "./src/app/denied-tokens.js",
        allowedIps: "./src/app/allowed-ips.js",
        endpointsConfig: "./src/app/endpoints-config.js",
        app: "./src/scss/app.scss",
      },
    },
    outDir: "../static",
  },
  plugins: [
    inject({
      include: "**/*.js", // Only include JavaScript files
      exclude: "**/*.scss", // Exclude SCSS files
      $: "jquery",
      jQuery: "jquery",
    }),
  ],
});
