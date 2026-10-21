const path = require("path");

/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
  experimental: { outputFileTracingRoot: path.join(__dirname, "../..") },
  reactStrictMode: true,
  transpilePackages: ["@hackways/ui"],
  poweredByHeader: false,
};

module.exports = nextConfig;
