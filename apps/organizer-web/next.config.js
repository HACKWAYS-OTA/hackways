const path = require("path");

/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
  basePath: "/manage",
  experimental: { outputFileTracingRoot: path.join(__dirname, "../..") },
  reactStrictMode: true,
  transpilePackages: ["@hackways/ui"],
  poweredByHeader: false,
  async headers() {
    return [{ source: "/auth/reset-password", headers: [{ key: "Referrer-Policy", value: "no-referrer" }, { key: "Cache-Control", value: "private, no-store" }] }];
  },
  async rewrites() {
    return [{ source: "/manage", destination: "/" }, { source: "/manage/", destination: "/" }];
  },
};

module.exports = nextConfig;
