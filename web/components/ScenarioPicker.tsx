"use client";

import type { ScenarioItem } from "@/lib/types";

interface ScenarioPickerProps {
  items: ScenarioItem[];
  selectedId?: string;
  loading?: boolean;
  onSelect: (scenario: ScenarioItem) => void;
  onStart: (scenario: ScenarioItem) => void;
}

export function ScenarioPicker({
  items,
  selectedId,
  loading,
  onSelect,
  onStart,
}: ScenarioPickerProps) {
  return (
    <section className="shrink-0">
      <div className="max-w-5xl mx-auto px-4 pt-2 pb-2">
        <div className="flex items-center justify-between gap-2 mb-2">
          <h2 className="text-sm font-semibold text-stone-800 font-serif">
            精选场景
          </h2>
          <p className="text-[11px] text-stone-500">一键填充议题与人物</p>
        </div>
        {loading && (
          <p className="text-xs text-stone-500 py-2">加载场景…</p>
        )}
        <div className="flex gap-2 overflow-x-auto pb-1">
          {items.map((s) => {
            const active = s.id === selectedId;
            return (
              <div
                key={s.id}
                className={`shrink-0 w-[220px] rounded-xl border p-3 transition ${
                  active
                    ? "border-stone-800 bg-stone-900 text-stone-50 shadow-sm"
                    : "border-stone-200 bg-white hover:border-stone-400"
                }`}
              >
                <button
                  type="button"
                  className="w-full text-left"
                  onClick={() => onSelect(s)}
                >
                  <p className="text-sm font-medium leading-snug">{s.title}</p>
                  <p
                    className={`text-[11px] mt-1.5 line-clamp-2 leading-relaxed ${
                      active ? "text-stone-300" : "text-stone-600"
                    }`}
                  >
                    {s.hook}
                  </p>
                  {s.expectedAngle && (
                    <p
                      className={`text-[10px] mt-2 ${
                        active ? "text-stone-400" : "text-stone-400"
                      }`}
                    >
                      视角：{s.expectedAngle}
                    </p>
                  )}
                </button>
                <button
                  type="button"
                  onClick={() => onStart(s)}
                  className={`mt-2 w-full text-[11px] py-1.5 rounded-lg border ${
                    active
                      ? "border-stone-600 text-stone-100 hover:bg-stone-800"
                      : "border-stone-300 hover:bg-stone-50"
                  }`}
                >
                  开始演示
                </button>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
