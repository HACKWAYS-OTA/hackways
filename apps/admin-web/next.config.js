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
    const core = (process.env.CORE_API_URL || "http://127.0.0.1:8081/api/v1").replace(/\/$/, "");
    const global = (process.env.GLOBAL_API_URL || "http://127.0.0.1:8082/api/v1").replace(/\/$/, "");
    return [
      { source: "/api/v1/admin/core/:path*", destination: `${core}/admin/core/:path*` },
      { source: "/api/v1/admin/global/:path*", destination: `${global}/admin/global/:path*` },
      { source: "/api/v1/auth/admin/:path*", destination: `${global}/auth/admin/:path*` },
    ];
  },
};

module.exports = nextConfig;
