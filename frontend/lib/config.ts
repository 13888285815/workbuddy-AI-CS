// 统一的 应用程序接口 配置
// 推荐生产模式：使用同域反向代理，将后端服务挂载在 /api 路径下。
// 这样可以避免在前端构建产物中硬编码主机名或端口。

export const API_BASE_URL = "";
export const API_PREFIX = "/api";

/**
 * 根据相对路径生成完整的 应用程序接口 地址
 * @param path 相对路径
 * @returns 完整的 应用程序接口 路径
 */
export function apiUrl(path: string): string {
  const p = path.startsWith("/") ? path : `/${path}`;
  return `${API_BASE_URL}${API_PREFIX}${p}`;
}

/** 
 * 获取管理端请求头
 * 包含当前用户标识与用于校验的身份令牌
 */
export function getAgentHeaders(): Record<string, string> {
  if (typeof window === "undefined") return {};
  const id = window.localStorage.getItem("agent_user_id");
  const token = window.localStorage.getItem("agent_ws_token");
  const headers: Record<string, string> = {};
  if (id) {
    headers["X-User-Id"] = id;
  }
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }
  return headers;
}
