"use client";

import type { CharacterItem } from "@/lib/types";

interface CharacterPickerProps {
  items: CharacterItem[];
  selectedId?: string;
  onSelect: (item: CharacterItem) => void;
}

export function CharacterPicker({
  items,
  selectedId,
  onSelect,
}: CharacterPickerProps) {
  return (
    <section className="space-y-3">
      <h2 className="text-sm font-semibold text-stone-800">选择历史人物</h2>
      <div className="grid gap-2 sm:grid-cols-1">
        {items.map((c) => {
          const active = c.id === selectedId;
          return (
            <button
              key={c.id}
              type="button"
              onClick={() => onSelect(c)}
              className={`rounded-lg border px-3 py-2 text-left transition ${
                active
                  ? "border-stone-800 bg-stone-900 text-stone-50"
                  : "border-stone-200 bg-white hover:border-stone-400"
              }`}
            >
              <div className="flex items-baseline justify-between gap-2">
                <span className="font-medium">{c.name}</span>
                <span className={`text-xs ${active ? "text-stone-300" : "text-stone-500"}`}>
                  {c.era}
                </span>
              </div>
              <p className={`text-xs mt-1 ${active ? "text-stone-300" : "text-stone-600"}`}>
                {c.summary}
              </p>
            </button>
          );
        })}
      </div>
    </section>
  );
}
