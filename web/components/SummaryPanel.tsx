"use client";

interface SummaryPanelProps {
  content?: string;
  loading?: boolean;
  error?: string;
  onGenerate: () => void;
  onCopy: () => void;
}

export function SummaryPanel({
  content,
  loading,
  error,
  onGenerate,
  onCopy,
}: SummaryPanelProps) {
  return (
    <section className="rounded-xl border border-stone-200 bg-white shadow-sm overflow-hidden">
      <header className="px-4 py-2.5 border-b border-stone-100 flex items-center justify-between gap-2 bg-stone-50/80">
        <p className="text-sm font-serif font-medium text-stone-800">圆桌摘要</p>
        <div className="flex gap-2">
          <button
            type="button"
            onClick={onGenerate}
            disabled={loading}
            className="text-[11px] px-2.5 py-1 rounded-lg bg-stone-900 text-white disabled:opacity-50"
          >
            {loading ? "生成中…" : content ? "重新生成" : "生成摘要"}
          </button>
          {content && (
            <button
              type="button"
              onClick={onCopy}
              className="text-[11px] px-2.5 py-1 rounded-lg border border-stone-300 hover:bg-white"
            >
              复制 Markdown
            </button>
          )}
        </div>
      </header>
      <div className="px-4 py-3 text-xs text-stone-700">
        {error && <p className="text-red-600 mb-2">{error}</p>}
        {!content && !loading && !error && (
          <p className="text-stone-500 leading-relaxed">
            群聊结束后，可生成结构化摘要，便于投稿知乎话题。
          </p>
        )}
        {content && (
          <pre className="whitespace-pre-wrap font-sans leading-relaxed text-stone-700">
            {content}
          </pre>
        )}
      </div>
    </section>
  );
}
