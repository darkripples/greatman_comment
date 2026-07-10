"use client";

import type { ConversationSummary } from "@/lib/types";

interface ConversationHistoryPanelProps {
  items: ConversationSummary[];
  activeId?: string;
  loading: boolean;
  error?: string;
  embedded?: boolean;
  onSelect: (item: ConversationSummary) => void;
  onRefresh: () => void;
}

function formatTime(ts: number) {
  if (!ts) return "";
  const d = new Date(ts * 1000);
  const now = new Date();
  const sameDay =
    d.getFullYear() === now.getFullYear() &&
    d.getMonth() === now.getMonth() &&
    d.getDate() === now.getDate();
  if (sameDay) {
    return d.toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" });
  }
  return d.toLocaleDateString("zh-CN", { month: "short", day: "numeric" });
}

function modeLabel(mode: string) {
  return mode === "group" ? "群聊" : "单人";
}

export function ConversationHistoryPanel({
  items,
  activeId,
  loading,
  error,
  embedded,
  onSelect,
  onRefresh,
}: ConversationHistoryPanelProps) {
  return (
    <section className="space-y-3">
      {!embedded && (
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-semibold text-stone-800">历史记录</h2>
          <button
            type="button"
            onClick={onRefresh}
            className="text-xs px-2 py-1 rounded-md border border-stone-300 hover:bg-stone-100"
          >
            刷新
          </button>
        </div>
      )}
      {embedded && (
        <div className="flex justify-end">
          <button
            type="button"
            onClick={onRefresh}
            className="text-xs px-2 py-1 rounded-md border border-stone-300 bg-white hover:bg-stone-50"
          >
            刷新
          </button>
        </div>
      )}
      {loading && <p className="text-xs text-stone-500">加载历史…</p>}
      {error && <p className="text-xs text-red-600">{error}</p>}
      <ul className={`space-y-2 pr-1 ${embedded ? "" : "max-h-[28vh] overflow-y-auto"}`}>
        {items.map((item) => {
          const active = item.id === activeId;
          const title = item.sourceTitle?.trim() || "未命名议题";
          return (
            <li key={item.id}>
              <button
                type="button"
                onClick={() => onSelect(item)}
                className={`w-full text-left rounded-lg border px-3 py-2 transition ${
                  active
                    ? "border-stone-700 bg-stone-100"
                    : "border-stone-200 bg-white hover:border-stone-300"
                }`}
              >
                <div className="flex items-center justify-between gap-2">
                  <span className="text-[10px] uppercase tracking-wide text-stone-500">
                    {modeLabel(item.mode)}
                  </span>
                  <span className="text-[10px] text-stone-400 shrink-0">
                    {formatTime(item.updatedAt)}
                  </span>
                </div>
                <p className="text-sm font-medium line-clamp-2 mt-0.5">{title}</p>
              </button>
            </li>
          );
        })}
        {items.length === 0 && !loading && !error && (
          <li className="text-xs text-stone-500">暂无历史对话，发送消息后会自动保存</li>
        )}
      </ul>
    </section>
  );
}
