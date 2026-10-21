const path = require("path");

/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
  experimental: { outputFileTracingRoot: path.join(__dirname, "../..") },
  reactStrictMode: true,
  transpilePackages: ["@hackways/ui"],
  poweredByHeader: false,
  async rewrites() {
    return [{ source: "/manage", destination: "/" }, { source: "/manage/", destination: "/" }];
  },
};

module.exports = nextConfig;
