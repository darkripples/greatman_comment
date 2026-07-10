"use client";

import type { AppSettings } from "@/lib/settings";
import { describeApiTarget } from "@/lib/settings";
import type { ProviderItem } from "@/lib/types";

interface SettingsPanelProps {
  settings: AppSettings;
  providers: ProviderItem[];
  apiStatus: string;
  hasZhihuKey?: boolean;
  hasDeepSeekKey?: boolean;
  onChange: (next: AppSettings) => void;
  onTestConnection: () => void;
}

export function SettingsPanel({
  settings,
  providers,
  apiStatus,
  hasZhihuKey,
  hasDeepSeekKey,
  onChange,
  onTestConnection,
}: SettingsPanelProps) {
  return (
    <section className="rounded-xl border border-stone-200 bg-stone-50 p-4 space-y-4">
      <div className="flex items-center justify-between gap-2">
        <h2 className="text-sm font-semibold text-stone-800">连接与模型</h2>
        <button
          type="button"
          onClick={onTestConnection}
          className="text-xs px-2 py-1 rounded-md border border-stone-300 hover:bg-white"
        >
          测试连接
        </button>
      </div>

      <p className="text-xs text-stone-500">
        设置保存在服务端 SQLite · {describeApiTarget(settings)}
      </p>
      <p className="text-xs text-stone-500">
        API Key 仅通过环境变量配置：
        {hasZhihuKey ? " 知乎已配置" : " 知乎未配置"}
        {hasDeepSeekKey ? " · DeepSeek 已配置" : " · DeepSeek 未配置"}
      </p>
      {apiStatus && (
        <p className="text-xs text-stone-600 bg-white border border-stone-200 rounded-md px-2 py-1">
          {apiStatus}
        </p>
      )}

      <div className="grid gap-3 sm:grid-cols-2">
        <label className="text-xs space-y-1">
          <span className="text-stone-600">API 环境</span>
          <select
            className="w-full rounded-md border border-stone-300 bg-white px-2 py-2 text-sm"
            value={settings.apiEnvironment}
            onChange={(e) =>
              onChange({
                ...settings,
                apiEnvironment: e.target.value as AppSettings["apiEnvironment"],
              })
            }
          >
            <option value="local">本地 Dev</option>
            <option value="prod">线上 Prod</option>
          </select>
        </label>

        {settings.apiEnvironment === "local" ? (
          <label className="text-xs space-y-1">
            <span className="text-stone-600">本地模式</span>
            <select
              className="w-full rounded-md border border-stone-300 bg-white px-2 py-2 text-sm"
              value={settings.localApiMode}
              onChange={(e) =>
                onChange({
                  ...settings,
                  localApiMode: e.target.value as AppSettings["localApiMode"],
                })
              }
            >
              <option value="rewrite">Next 反代 (/api)</option>
              <option value="direct">直连 Go</option>
            </select>
          </label>
        ) : (
          <label className="text-xs space-y-1 sm:col-span-1">
            <span className="text-stone-600">线上 API 根地址</span>
            <input
              className="w-full rounded-md border border-stone-300 bg-white px-2 py-2 text-sm"
              value={settings.prodApiBase}
              onChange={(e) =>
                onChange({ ...settings, prodApiBase: e.target.value })
              }
              placeholder="https://api.example.com"
            />
          </label>
        )}
      </div>

      {settings.apiEnvironment === "local" && settings.localApiMode === "direct" && (
        <label className="text-xs space-y-1 block">
          <span className="text-stone-600">本地 Go 地址</span>
          <input
            className="w-full rounded-md border border-stone-300 bg-white px-2 py-2 text-sm"
            value={settings.devApiBase}
            onChange={(e) => onChange({ ...settings, devApiBase: e.target.value })}
          />
        </label>
      )}

      <label className="text-xs space-y-1 block">
        <span className="text-stone-600">LLM 提供方</span>
        <select
          className="w-full rounded-md border border-stone-300 bg-white px-2 py-2 text-sm"
          value={settings.llmProvider}
          onChange={(e) =>
            onChange({
              ...settings,
              llmProvider: e.target.value as AppSettings["llmProvider"],
            })
          }
        >
          {providers.length > 0 ? (
            providers.map((p) => (
              <option key={p.id} value={p.id} disabled={!p.available}>
                {p.name} ({p.model}){p.available ? "" : " · 未配置 Key"}
              </option>
            ))
          ) : (
            <>
              <option value="deepseek">DeepSeek</option>
              <option value="zhihu">知乎直答</option>
            </>
          )}
        </select>
      </label>

      <details className="text-xs">
        <summary className="cursor-pointer text-stone-700 font-medium">高级 · 缓存与模型</summary>
        <div className="mt-3 space-y-3 pt-3 border-t border-stone-200">
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={settings.zhihuMock}
              onChange={(e) => onChange({ ...settings, zhihuMock: e.target.checked })}
            />
            <span>知乎 Mock（热榜/搜索用本地 fixtures）</span>
          </label>
          <div className="grid gap-2 sm:grid-cols-2">
            <label className="space-y-1">
              <span className="text-stone-600">热榜缓存 TTL</span>
              <input
                className="w-full rounded-md border border-stone-300 bg-white px-2 py-1.5"
                value={settings.hotListCacheTtl}
                onChange={(e) => onChange({ ...settings, hotListCacheTtl: e.target.value })}
              />
            </label>
            <label className="space-y-1">
              <span className="text-stone-600">热榜最小拉取间隔</span>
              <input
                className="w-full rounded-md border border-stone-300 bg-white px-2 py-1.5"
                value={settings.hotListMinInterval}
                onChange={(e) => onChange({ ...settings, hotListMinInterval: e.target.value })}
              />
            </label>
          </div>
          <div className="grid gap-2 sm:grid-cols-2">
            <label className="space-y-1">
              <span className="text-stone-600">DeepSeek 模型</span>
              <input
                className="w-full rounded-md border border-stone-300 bg-white px-2 py-1.5"
                value={settings.deepseekModel}
                onChange={(e) => onChange({ ...settings, deepseekModel: e.target.value })}
              />
            </label>
            <label className="space-y-1">
              <span className="text-stone-600">知乎直答模型</span>
              <input
                className="w-full rounded-md border border-stone-300 bg-white px-2 py-1.5"
                value={settings.zhidaModel}
                onChange={(e) => onChange({ ...settings, zhidaModel: e.target.value })}
              />
            </label>
          </div>
        </div>
      </details>
    </section>
  );
}
