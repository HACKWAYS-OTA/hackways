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
};

module.exports = nextConfig;
