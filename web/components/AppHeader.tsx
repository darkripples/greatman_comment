"use client";

import { ModeToggle, type ChatMode } from "./ModeToggle";

interface AppHeaderProps {
  chatMode: ChatMode;
  onChatModeChange: (mode: ChatMode) => void;
  onOpenSettings: () => void;
  onOpenHistory: () => void;
  historyCount?: number;
}

export function AppHeader({
  chatMode,
  onChatModeChange,
  onOpenSettings,
  onOpenHistory,
  historyCount,
}: AppHeaderProps) {
  return (
    <header className="shrink-0 border-b border-stone-200 bg-white/95 backdrop-blur z-20">
      <div className="max-w-5xl mx-auto px-4 py-3 flex items-center justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-base md:text-lg font-serif font-semibold text-stone-900 truncate">
            用 AI 重新看见人
          </h1>
          <p className="text-[11px] text-stone-500">人文季 · 历史单元</p>
        </div>
        <div className="flex items-center gap-2 shrink-0">
          <ModeToggle mode={chatMode} onChange={onChatModeChange} compact />
          <button
            type="button"
            onClick={onOpenHistory}
            className="relative text-xs px-3 py-2 rounded-lg border border-stone-300 hover:bg-stone-50"
          >
            历史
            {historyCount != null && historyCount > 0 && (
              <span className="absolute -top-1.5 -right-1.5 min-w-[18px] h-[18px] px-1 rounded-full bg-stone-800 text-white text-[10px] leading-[18px] text-center">
                {historyCount > 99 ? "99+" : historyCount}
              </span>
            )}
          </button>
          <button
            type="button"
            onClick={onOpenSettings}
            className="text-xs px-3 py-2 rounded-lg bg-stone-900 text-white hover:bg-stone-800"
          >
            设置
          </button>
        </div>
      </div>
    </header>
  );
}
