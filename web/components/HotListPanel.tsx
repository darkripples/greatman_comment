"use client";

import { useCallback, useEffect, useState } from "react";
import { fetchSearch } from "@/lib/api";
import type { AppSettings } from "@/lib/settings";
import type { HotItem } from "@/lib/types";

interface HotListPanelProps {
  settings: AppSettings;
  items: HotItem[];
  selectedTitle?: string;
  loading: boolean;
  error?: string;
  onSelect: (item: HotItem) => void;
  onRefresh: () => void;
}

function toHotItem(item: { title: string; url: string; excerpt?: string }): HotItem {
  return { title: item.title, url: item.url, excerpt: item.excerpt };
}

export function HotListPanel({
  settings,
  items,
  selectedTitle,
  loading,
  error,
  onSelect,
  onRefresh,
}: HotListPanelProps) {
  const [query, setQuery] = useState("");
  const [searchItems, setSearchItems] = useState<HotItem[]>([]);
  const [searchLoading, setSearchLoading] = useState(false);
  const [searchError, setSearchError] = useState<string>();
  const [mode, setMode] = useState<"hot" | "search">("hot");

  const runSearch = useCallback(
    async (q: string) => {
      const trimmed = q.trim();
      if (!trimmed) {
        setSearchItems([]);
        setMode("hot");
        return;
      }
      setMode("search");
      setSearchLoading(true);
      setSearchError(undefined);
      try {
        const results = await fetchSearch(settings, trimmed);
        setSearchItems(results.map(toHotItem));
      } catch (e) {
        setSearchError(e instanceof Error ? e.message : "搜索失败");
        setSearchItems([]);
      } finally {
        setSearchLoading(false);
      }
    },
    [settings],
  );

  useEffect(() => {
    const t = setTimeout(() => {
      void runSearch(query);
    }, 300);
    return () => clearTimeout(t);
  }, [query, runSearch]);

  const displayItems = mode === "search" ? searchItems : items;
  const listLoading = mode === "search" ? searchLoading : loading;
  const listError = mode === "search" ? searchError : error;

  return (
    <section className="shrink-0">
      <div className="max-w-5xl mx-auto px-4 pt-3 pb-2">
        <div className="flex items-center justify-between gap-3 mb-2">
          <div>
            <h2 className="text-sm font-semibold text-stone-800">
              {mode === "search" ? "搜索议题" : "今日之问 · 知乎热榜"}
            </h2>
            <p className="text-[11px] text-stone-500 mt-0.5">
              选一条议题，或搜索关键词开启对话
            </p>
          </div>
          <button
            type="button"
            onClick={onRefresh}
            disabled={loading}
            className="text-xs px-2.5 py-1.5 rounded-md border border-stone-300 bg-white hover:bg-stone-50 disabled:opacity-50"
          >
            {loading ? "加载中…" : "刷新热榜"}
          </button>
        </div>

        <div className="mb-2">
          <input
            type="search"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="搜索知乎议题…"
            className="w-full text-sm rounded-lg border border-stone-300 px-3 py-2 bg-white focus:outline-none focus:ring-2 focus:ring-amber-600/25"
          />
        </div>

        {listError && (
          <p className="text-xs text-red-600 mb-2 px-1">{listError}</p>
        )}

        <div className="flex gap-3 overflow-x-auto pb-1 -mx-1 px-1 snap-x snap-mandatory scrollbar-thin">
          {(displayItems ?? []).map((item, index) => {
            const active = item.title === selectedTitle;
            return (
              <button
                key={`${item.title}-${index}`}
                type="button"
                onClick={() => onSelect(item)}
                className={`snap-start shrink-0 w-[min(280px,78vw)] text-left rounded-xl border px-3 py-2.5 transition ${
                  active
                    ? "border-amber-600 bg-amber-50 shadow-sm ring-1 ring-amber-600/20"
                    : "border-stone-200 bg-white hover:border-stone-300 hover:shadow-sm"
                }`}
              >
                <div className="flex items-start gap-2">
                  <span
                    className={`font-serif text-lg leading-none pt-0.5 ${
                      active ? "text-amber-700" : "text-stone-400"
                    }`}
                  >
                    {index + 1}
                  </span>
                  <div className="min-w-0 flex-1">
                    <p className="text-sm font-medium line-clamp-2 leading-snug">
                      {item.title}
                    </p>
                    {item.excerpt && (
                      <p className="text-[11px] text-stone-500 mt-1 line-clamp-2 leading-relaxed">
                        {item.excerpt}
                      </p>
                    )}
                  </div>
                </div>
              </button>
            );
          })}
          {(displayItems ?? []).length === 0 && !listLoading && !listError && (
            <p className="text-xs text-stone-500 py-4 px-1">
              {mode === "search" ? "无搜索结果" : "暂无热榜数据"}
            </p>
          )}
          {listLoading && (
            <p className="text-xs text-stone-500 py-4 px-1 animate-pulse">加载中…</p>
          )}
        </div>
      </div>
    </section>
  );
}
