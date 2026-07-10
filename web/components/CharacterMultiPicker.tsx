"use client";

import type { CharacterItem } from "@/lib/types";
import { MAX_GROUP_MEMBERS } from "@/lib/types";

interface CharacterMultiPickerProps {
  items: CharacterItem[];
  selectedIds: string[];
  onChange: (ids: string[]) => void;
}

export function CharacterMultiPicker({
  items,
  selectedIds,
  onChange,
}: CharacterMultiPickerProps) {
  const toggle = (id: string) => {
    if (selectedIds.includes(id)) {
      onChange(selectedIds.filter((x) => x !== id));
      return;
    }
    if (selectedIds.length >= MAX_GROUP_MEMBERS) return;
    onChange([...selectedIds, id]);
  };

  const names = items
    .filter((c) => selectedIds.includes(c.id))
    .map((c) => c.name)
    .join("、");

  return (
    <section className="space-y-3">
      <div>
        <h2 className="text-sm font-semibold text-stone-800">选择群聊成员</h2>
        <p className="text-xs text-stone-500 mt-1">
          至少 2 人，最多 {MAX_GROUP_MEMBERS} 人{names ? ` · 已选：${names}` : ""}
        </p>
      </div>
      <div className="grid gap-2">
        {items.map((c) => {
          const checked = selectedIds.includes(c.id);
          return (
            <label
              key={c.id}
              className={`flex gap-3 rounded-lg border px-3 py-2 cursor-pointer ${
                checked ? "border-amber-700 bg-amber-50" : "border-stone-200 bg-white"
              }`}
            >
              <input
                type="checkbox"
                checked={checked}
                onChange={() => toggle(c.id)}
                className="mt-1"
              />
              <span className="flex-1">
                <span className="font-medium text-sm">{c.name}</span>
                <span className="text-xs text-stone-500 ml-2">{c.era}</span>
                <p className="text-xs text-stone-600 mt-1">{c.summary}</p>
              </span>
            </label>
          );
        })}
      </div>
    </section>
  );
}
