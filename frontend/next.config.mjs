/**
 * 生产环境适用的 网页应用配置
 */

// 开发环境下代理后端服务的端口（统一从根目录环境变量文件中读取）
const backendPort = process.env.NEXT_PUBLIC_BACKEND_PORT || "18080";
const backendHost = process.env.NEXT_PUBLIC_BACKEND_HOST || "localhost";

/** @type {import('next').NextConfig} */
const nextConfig = {
  async rewrites() {
    // 仅在开发模式下启用内部请求转发逻辑
    if (process.env.NODE_ENV === "development") {
      return [
        // 将所有以 /api 开头的请求转发至后端服务
        {
          source: "/api/:path*",
          destination: `http://${backendHost}:${backendPort}/api/:path*`,
        },
        // 转发用户个人资料相关请求
        {
          source: "/agent/profile/:path*",
          destination: `http://${backendHost}:${backendPort}/agent/profile/:path*`,
        },
        // 转发头像上传与读取请求
        {
          source: "/agent/avatar/:path*",
          destination: `http://${backendHost}:${backendPort}/agent/avatar/:path*`,
        },
        // 转发嵌入服务配置请求
        {
          source: "/agent/embedding-config",
          destination: `http://${backendHost}:${backendPort}/agent/embedding-config`,
        },
        // 转发人工智能配置请求
        {
          source: "/agent/ai-config/:path*",
          destination: `http://${backendHost}:${backendPort}/agent/ai-config/:path*`,
        },
        // 转发数据统计摘要请求
        {
          source: "/agent/analytics/summary",
          destination: `http://${backendHost}:${backendPort}/agent/analytics/summary`,
        },
        // 转发后端系统日志请求
        {
          source: "/agent/logs/api",
          destination: `http://${backendHost}:${backendPort}/agent/logs/api`,
        },
        // 转发前端日志上报请求
        {
          source: "/agent/logs/frontend",
          destination: `http://${backendHost}:${backendPort}/agent/logs/frontend`,
        },
        // 转发日志等级配置请求
        {
          source: "/agent/logs/min-level",
          destination: `http://${backendHost}:${backendPort}/agent/logs/min-level`,
        },
        // 转发其他所有未匹配的非静态资源请求至后端
        {
          source: "/:path((?!_next|agent|chat|favicon.ico).*)",
          destination: `http://${backendHost}:${backendPort}/:path*`,
        },
      ];
    }
    return [];
  },
  images: {
    // 允许加载来自以下地址的远程图片资源（如上传的头像）
    remotePatterns: [
      {
        protocol: "http",
        hostname: "localhost",
        port: backendPort,
        pathname: "/uploads/**",
      },
      {
        protocol: "http",
        hostname: "127.0.0.1",
        port: backendPort,
        pathname: "/uploads/**",
      },
      {
        // 匹配任意主机名，以增强兼容性
        protocol: "http",
        hostname: "**",
        port: backendPort,
        pathname: "/uploads/**",
      },
    ],
  },
};

export default nextConfig;
