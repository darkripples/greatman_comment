"use client";

import type { HotItem } from "@/lib/types";

interface TopicBriefPanelProps {
  item?: HotItem;
}

const PROMPT_HINTS = [
  "今日议题往往折射时代焦虑——古人会以何种价值尺度回应？",
  "可追问：若受时代边界所限，哪些判断必须保留，哪些必须悬置？",
  "群聊模式下，不同人物的分歧本身即是答案的一部分。",
];

export function TopicBriefPanel({ item }: TopicBriefPanelProps) {
  if (!item?.title) return null;

  return (
    <div className="max-w-5xl mx-auto px-4 pb-2">
      <div className="rounded-xl border border-amber-200/80 bg-amber-50/60 px-3 py-2.5 text-xs text-stone-700">
        <p className="font-medium text-stone-800 font-serif">{item.title}</p>
        {item.excerpt && (
          <p className="mt-1 leading-relaxed text-stone-600">{item.excerpt}</p>
        )}
        {item.detail_text && (
          <p className="mt-1 leading-relaxed text-stone-500 line-clamp-3">
            {item.detail_text}
          </p>
        )}
        {item.url && (
          <a
            href={item.url}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-block mt-1.5 text-amber-800 underline underline-offset-2 hover:text-amber-900"
          >
            查看原文
          </a>
        )}
        <div className="mt-2 pt-2 border-t border-amber-200/60">
          <p className="text-[11px] font-medium text-stone-700 mb-1">为什么值得问古人？</p>
          <ul className="space-y-0.5 text-[11px] text-stone-600">
            {PROMPT_HINTS.map((h) => (
              <li key={h}>· {h}</li>
            ))}
          </ul>
        </div>
      </div>
    </div>
  );
}
