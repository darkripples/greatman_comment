export type ApiEnvironment = "local" | "prod";
export type LocalApiMode = "rewrite" | "direct";
export type LLMProviderId = "deepseek" | "zhihu";
export type ChatModeId = "single" | "group";

/** 与后端 SQLite app_settings 对齐 */
export interface AppSettings {
  apiEnvironment: ApiEnvironment;
  localApiMode: LocalApiMode;
  prodApiBase: string;
  devApiBase: string;
  llmProvider: LLMProviderId;
  deepseekModel: string;
  zhidaModel: string;
  deepseekApiBase: string;
  zhihuApiBase: string;
  zhihuMock: boolean;
  hotListCacheTtl: string;
  hotListMinInterval: string;
  searchCacheTtl: string;
  searchMinInterval: string;
  chatMode: ChatModeId;
}

export interface AppSettingsResponse extends AppSettings {
  hasZhihuKey?: boolean;
  hasDeepSeekKey?: boolean;
  source?: string;
}

export const defaultSettings: AppSettings = {
  apiEnvironment: "local",
  localApiMode: "rewrite",
  prodApiBase: "https://your-prod-api.example.com",
  devApiBase: "http://127.0.0.1:30302",
  llmProvider: "deepseek",
  deepseekModel: "deepseek-v4-flash",
  zhidaModel: "zhida-fast-1p5",
  deepseekApiBase: "https://api.deepseek.com",
  zhihuApiBase: "https://developer.zhihu.com",
  zhihuMock: false,
  hotListCacheTtl: "5m",
  hotListMinInterval: "5h",
  searchCacheTtl: "5m",
  searchMinInterval: "5m",
  chatMode: "single",
};

export function getApiBase(settings: AppSettings): string {
  if (settings.apiEnvironment === "prod") {
    return (settings.prodApiBase || defaultSettings.prodApiBase).replace(/\/$/, "");
  }
  if (settings.localApiMode === "direct") {
    return (settings.devApiBase || defaultSettings.devApiBase).replace(/\/$/, "");
  }
  return "";
}

export function describeApiTarget(settings: AppSettings): string {
  const base = getApiBase(settings);
  if (settings.apiEnvironment === "local" && settings.localApiMode === "rewrite") {
    return "本地 · Next 反代 (/api → Go)";
  }
  return `${settings.apiEnvironment === "local" ? "本地 · 直连" : "线上"} · ${base || "(同源)"}`;
}

export function toApiSettingsPayload(settings: AppSettings): AppSettings {
  return { ...settings };
}
