/**
 * 官网配置文件
 * 请根据实际情况修改以下配置
 */

export const websiteConfig = {
  /** 在线演示站点（「立即体验 / 打开 演示」等入口） */
  demoUrl: "https://tools.yndxw.com",

  // 代码托管平台仓库地址
  github: {
    repo: "https://github.com/yndxw/workbuddy-ai",
    releases: "https://github.com/yndxw/workbuddy-ai/releases",
    issues: "https://github.com/yndxw/workbuddy-ai/issues",
    readme: "https://github.com/yndxw/workbuddy-ai/blob/master/README.md",
  },
  
  // 联系方式
  contact: {
    email: "zzx@yndxw.com", // 可选：电子邮箱地址
    wechat: "", // 可选：微信号或微信群链接
  },
  
  // 友情链接（用于互相引流）
  // 格式：{ name: "链接名称", url: "链接地址" }
  friendLinks: [
    // { name: "示例链接", url: "https://example.com" },
  ] as Array<{ name: string; url: string }>,
  
  // 其他配置
  copyright: {
    company: "WorkBuddy AI 智能客服系统", // 公司或产品名称
    year: new Date().getFullYear(),
  },
};
