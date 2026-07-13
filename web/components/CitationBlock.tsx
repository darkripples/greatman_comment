"use client";

import type { Citation } from "@/lib/types";

interface CitationBlockProps {
  citations?: Citation[];
}

export function CitationBlock({ citations }: CitationBlockProps) {
  if (!citations || citations.length === 0) return null;

  return (
    <details className="mt-2 rounded-lg border border-stone-200 bg-stone-50/80 text-xs">
      <summary className="cursor-pointer px-2.5 py-1.5 text-stone-600 hover:text-stone-800">
        参考史料（{citations.length}）
      </summary>
      <ul className="px-2.5 pb-2 space-y-1.5 border-t border-stone-200/80">
        {citations.map((c, i) => (
          <li key={`${c.title}-${i}`} className="text-stone-600 leading-relaxed">
            {c.title && <span className="font-medium text-stone-700">{c.title}</span>}
            {c.source && (
              <span className="text-stone-500"> · {c.source}</span>
            )}
            {c.excerpt && (
              <p className="text-[11px] text-stone-500 mt-0.5 line-clamp-3">{c.excerpt}</p>
            )}
          </li>
        ))}
      </ul>
    </details>
  );
}
