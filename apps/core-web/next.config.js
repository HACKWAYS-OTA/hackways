const path = require("path");

/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
  // Separate static chunks from public-web on the shared production origin.
  assetPrefix: process.env.NODE_ENV === "production" ? "/_core" : undefined,
  experimental: { outputFileTracingRoot: path.join(__dirname, "../..") },
  reactStrictMode: true,
  transpilePackages: ["@hackways/ui"],
  poweredByHeader: false,
  async rewrites() {
    if (process.env.NODE_ENV !== 'development') return [];
    const global = (process.env.GLOBAL_API_URL || 'http://127.0.0.1:8082/api/v1').replace(/\/$/, '');
    return [{ source: '/api/v1/auth/:path*', destination: `${global}/auth/:path*` }];
  },
  async headers() {
    return [{ source: "/auth/unsubscribe", headers: [{ key: "Referrer-Policy", value: "no-referrer" }, { key: "Cache-Control", value: "private, no-store" }] }, { source: "/auth/reset-password", headers: [{ key: "Referrer-Policy", value: "no-referrer" }, { key: "Cache-Control", value: "private, no-store" }] }];
  },
};

module.exports = nextConfig;
