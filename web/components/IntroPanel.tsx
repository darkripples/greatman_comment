"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

const STORAGE_KEY = "renwen_intro_collapsed";

interface IntroPanelProps {
  onWatchDemo: () => void;
  demoLoading?: boolean;
}

export function IntroPanel({ onWatchDemo, demoLoading }: IntroPanelProps) {
  const [collapsed, setCollapsed] = useState(false);

  useEffect(() => {
    try {
      setCollapsed(localStorage.getItem(STORAGE_KEY) === "1");
    } catch {
      /* ignore */
    }
  }, []);

  const toggle = () => {
    setCollapsed((v) => {
      const next = !v;
      try {
        localStorage.setItem(STORAGE_KEY, next ? "1" : "0");
      } catch {
        /* ignore */
      }
      return next;
    });
  };

  if (collapsed) {
    return (
      <div className="max-w-5xl mx-auto px-4 pt-3 pb-1">
        <button
          type="button"
          onClick={toggle}
          className="text-xs text-stone-600 hover:text-stone-900 underline underline-offset-2"
        >
          展开作品说明 · 把今日之问，交给昨日之人
        </button>
      </div>
    );
  }

  return (
    <section className="max-w-5xl mx-auto px-4 pt-3 pb-2">
      <div className="rounded-2xl border border-stone-200 bg-white/90 shadow-sm overflow-hidden">
        <div className="px-4 py-3 border-b border-stone-100 flex items-start justify-between gap-3">
          <div>
            <h2 className="text-base font-serif font-semibold text-stone-900">
              把今日之问，交给昨日之人
            </h2>
            <p className="text-xs text-stone-500 mt-1 leading-relaxed">
              有限视角下的跨时空对话 · 非穿越爽文 · 人文季历史单元
            </p>
          </div>
          <button
            type="button"
            onClick={toggle}
            className="shrink-0 text-xs text-stone-500 hover:text-stone-800 px-2 py-1"
          >
            收起
          </button>
        </div>
        <div className="px-4 py-3 grid md:grid-cols-3 gap-3 text-xs text-stone-700">
          <div className="rounded-xl bg-stone-50 p-3 border border-stone-100">
            <p className="font-medium text-stone-800 mb-1">时代边界</p>
            <p className="leading-relaxed text-stone-600">
              人物不装全知。超出其时代的问题，会在正文中自然点明局限。
            </p>
          </div>
          <div className="rounded-xl bg-stone-50 p-3 border border-stone-100">
            <p className="font-medium text-stone-800 mb-1">史料优先</p>
            <p className="leading-relaxed text-stone-600">
              基于可考思想立场与语料片段回应，而非随意编造史实。
            </p>
          </div>
          <div className="rounded-xl bg-stone-50 p-3 border border-stone-100">
            <p className="font-medium text-stone-800 mb-1">群聊碰撞</p>
            <p className="leading-relaxed text-stone-600">
              不同价值观在同一议题下辩论，形成多视角圆桌。
            </p>
          </div>
        </div>
        <div className="px-4 pb-3 flex flex-wrap items-center gap-2">
          <button
            type="button"
            onClick={onWatchDemo}
            disabled={demoLoading}
            className="text-xs px-3 py-2 rounded-lg bg-amber-800 text-white hover:bg-amber-900 disabled:opacity-50"
          >
            {demoLoading ? "加载演示…" : "观看精选演示"}
          </button>
          <Link
            href="/about"
            className="text-xs px-3 py-2 rounded-lg border border-stone-300 hover:bg-stone-50"
          >
            作品说明
          </Link>
          <p className="text-[11px] text-stone-500 ml-auto">
            边界：AI 生成的是「文本人格」，非史实复原
          </p>
        </div>
      </div>
    </section>
  );
}
