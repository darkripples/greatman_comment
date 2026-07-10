import type { NextConfig } from "next";

const devApiOrigin = process.env.DEV_API_ORIGIN || "http://127.0.0.1:30302";
const proxyTimeoutMs = Number(process.env.DEV_API_PROXY_TIMEOUT_MS || 180_000);

const nextConfig: NextConfig = {
  experimental: {
    // Next rewrites 默认 30s，LLM 对话/群聊需更长
    proxyTimeout: proxyTimeoutMs,
  },
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${devApiOrigin}/api/:path*`,
      },
    ];
  },
};

export default nextConfig;
