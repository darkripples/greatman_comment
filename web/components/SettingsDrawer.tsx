"use client";

import type { AppSettings } from "@/lib/settings";
import type { ProviderItem } from "@/lib/types";
import { SettingsPanel } from "./SettingsPanel";

interface SettingsDrawerProps {
  open: boolean;
  onClose: () => void;
  settings: AppSettings;
  providers: ProviderItem[];
  apiStatus: string;
  hasZhihuKey?: boolean;
  hasDeepSeekKey?: boolean;
  onSettingsChange: (next: AppSettings) => void;
  onTestConnection: () => void;
}

export function SettingsDrawer({
  open,
  onClose,
  settings,
  providers,
  apiStatus,
  hasZhihuKey,
  hasDeepSeekKey,
  onSettingsChange,
  onTestConnection,
}: SettingsDrawerProps) {
  if (!open) return null;

  return (
    <div className="fixed inset-0 z-40">
      <button
        type="button"
        className="absolute inset-0 bg-black/35 backdrop-blur-[1px]"
        aria-label="关闭"
        onClick={onClose}
      />
      <aside className="absolute inset-y-0 right-0 w-full max-w-md bg-[#f8f6f2] shadow-2xl flex flex-col border-l border-stone-200">
        <header className="flex items-center justify-between gap-3 px-4 py-3 border-b border-stone-200 bg-white">
          <h2 className="font-medium text-stone-900">设置</h2>
          <button
            type="button"
            onClick={onClose}
            className="text-sm px-3 py-1.5 rounded-lg border border-stone-300 hover:bg-stone-50"
          >
            关闭
          </button>
        </header>
        <div className="flex-1 overflow-y-auto p-4">
          <SettingsPanel
            settings={settings}
            providers={providers}
            apiStatus={apiStatus}
            hasZhihuKey={hasZhihuKey}
            hasDeepSeekKey={hasDeepSeekKey}
            onChange={onSettingsChange}
            onTestConnection={onTestConnection}
          />
        </div>
      </aside>
    </div>
  );
}
