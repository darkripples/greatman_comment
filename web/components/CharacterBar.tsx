"use client";

import type { CharacterItem } from "@/lib/types";
import { MAX_GROUP_MEMBERS, characterAccent } from "@/lib/types";
import type { ChatMode } from "./ModeToggle";

interface CharacterBarProps {
  mode: ChatMode;
  items: CharacterItem[];
  selectedId?: string;
  selectedGroupIds: string[];
  onSelectSingle: (item: CharacterItem) => void;
  onToggleGroup: (id: string) => void;
}

export function CharacterBar({
  mode,
  items,
  selectedId,
  selectedGroupIds,
  onSelectSingle,
  onToggleGroup,
}: CharacterBarProps) {
  const toggleGroup = (id: string) => {
    if (selectedGroupIds.includes(id)) {
      if (selectedGroupIds.length <= 2) return;
      onToggleGroup(id);
      return;
    }
    if (selectedGroupIds.length >= MAX_GROUP_MEMBERS) return;
    onToggleGroup(id);
  };

  return (
    <section className="shrink-0">
      <div className="max-w-5xl mx-auto px-4 pt-2 pb-2.5 border-t border-stone-200/50">
        <div className="flex items-center justify-between gap-2 mb-2">
          <h2 className="text-sm font-semibold text-stone-800">
            {mode === "group" ? "群聊成员" : "对话人物"}
          </h2>
          {mode === "group" && (
            <p className="text-[11px] text-stone-500">
              已选 {selectedGroupIds.length}/{MAX_GROUP_MEMBERS} · 至少 2 人
            </p>
          )}
        </div>
        <div className="flex gap-2 overflow-x-auto pb-0.5">
          {items.map((c) => {
            const active =
              mode === "group"
                ? selectedGroupIds.includes(c.id)
                : c.id === selectedId;
            return (
              <button
                key={c.id}
                type="button"
                onClick={() =>
                  mode === "group" ? toggleGroup(c.id) : onSelectSingle(c)
                }
                className={`shrink-0 min-w-[140px] max-w-[200px] text-left rounded-xl border px-3 py-2 transition border-l-4 ${characterAccent(c.id)} ${
                  active
                    ? "border-stone-800 bg-stone-900 text-stone-50 shadow-sm"
                    : "border-stone-200 bg-[#faf8f5] hover:border-stone-400 hover:bg-white"
                }`}
              >
                <div className="flex items-baseline justify-between gap-1">
                  <span className="text-sm font-medium">{c.name}</span>
                  <span
                    className={`text-[10px] ${active ? "text-stone-400" : "text-stone-500"}`}
                  >
                    {c.era}
                  </span>
                </div>
                <p
                  className={`text-[11px] mt-0.5 line-clamp-1 ${active ? "text-stone-300" : "text-stone-600"}`}
                >
                  {c.summary}
                </p>
              </button>
            );
          })}
          {items.length === 0 && (
            <p className="text-xs text-stone-500 py-2">加载人物列表…</p>
          )}
        </div>
      </div>
    </section>
  );
}
