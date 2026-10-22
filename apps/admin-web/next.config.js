const path = require("path");

/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
  experimental: { outputFileTracingRoot: path.join(__dirname, "../..") },
  reactStrictMode: true,
  transpilePackages: ["@hackways/ui"],
  poweredByHeader: false,
  async rewrites() {
    if (process.env.NODE_ENV !== "development") return [];
    return [
      { source: "/api/v1/admin/core/:path*", destination: "http://127.0.0.1:8081/api/v1/admin/core/:path*" },
      { source: "/api/v1/admin/global/:path*", destination: "http://127.0.0.1:8082/api/v1/admin/global/:path*" },
      { source: "/api/v1/auth/admin/:path*", destination: "http://127.0.0.1:8082/api/v1/auth/admin/:path*" },
    ];
  },
};

module.exports = nextConfig;
