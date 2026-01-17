import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import path from "path";
import { dynamicBase } from "vite-plugin-dynamic-base";

export default defineConfig({
    base: process.env.NODE_ENV === "production" ? "/__dynamic_base__/" : (process.env.VITE_CLOUD_SENTINEL_K8S_BASE || "/"),
    plugins: [
        dynamicBase({
            publicPath: "window.__dynamic_base__",
            transformIndexHtml: true,
        }),
        react(),
        tailwindcss(),
    ],
    envPrefix: ["VITE_", "CLOUD_SENTINEL_K8S_"],
    resolve: {
        alias: {
            "@": path.resolve(__dirname, "./src"),
        },
    },
    server: {
        port: 3000,
    },
});
