"use client";

import type { ChatMessage, CharacterItem } from "@/lib/types";
import { MAX_GROUP_MEMBERS, characterAccent } from "@/lib/types";
import type { ChatMode } from "./ModeToggle";

interface ChatWindowProps {
  mode: ChatMode;
  messages: ChatMessage[];
  input: string;
  loading: boolean;
  loadingHint?: string;
  characterName?: string;
  groupCharacters?: CharacterItem[];
  sourceTitle?: string;
  canSend: boolean;
  providerHint?: string;
  onInputChange: (value: string) => void;
  onSend: () => void;
  onNewDiscussion: () => void;
}

export function ChatWindow({
  mode,
  messages,
  input,
  loading,
  loadingHint,
  characterName,
  groupCharacters,
  sourceTitle,
  canSend,
  providerHint,
  onInputChange,
  onSend,
  onNewDiscussion,
}: ChatWindowProps) {
  const participants =
    mode === "group"
      ? groupCharacters?.map((c) => c.name).join("、") || "未选人物"
      : characterName || "未选人物";

  return (
    <section className="flex flex-col h-full min-h-0 rounded-2xl border border-stone-200 bg-white shadow-sm overflow-hidden">
      <header className="px-4 py-3 border-b border-stone-200 bg-stone-50/80 flex flex-wrap items-start justify-between gap-2">
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <p className="text-sm font-medium text-stone-800">
              {mode === "group" ? "群聊讨论" : "对话"}
            </p>
            <span className="text-[11px] px-2 py-0.5 rounded-full bg-stone-200/80 text-stone-700">
              {participants}
            </span>
          </div>
          {sourceTitle ? (
            <p className="text-xs text-stone-500 mt-1.5 line-clamp-2 leading-relaxed">
              议题：{sourceTitle}
            </p>
          ) : (
            <p className="text-xs text-stone-400 mt-1.5">
              在上方选择人物与热榜议题，或直接输入问题
            </p>
          )}
        </div>
        <button
          type="button"
          onClick={onNewDiscussion}
          className="shrink-0 text-xs px-2.5 py-1.5 border border-stone-300 rounded-lg hover:bg-white bg-white/80"
        >
          新讨论
        </button>
      </header>

      <div className="flex-1 min-h-0 overflow-y-auto px-4 py-4 space-y-3 bg-[#faf8f5]">
        {messages.length === 0 && (
          <div className="h-full flex flex-col items-center justify-center text-center px-6 py-8">
            <p className="text-sm text-stone-600 font-serif">跨越时空的对话</p>
            <p className="text-xs text-stone-500 mt-2 max-w-sm leading-relaxed">
              {mode === "group"
                ? `在上方选择 2–${MAX_GROUP_MEMBERS} 位群聊成员，点击热榜议题或输入问题开始。`
                : "在上方选择一位历史人物，点击热榜议题或输入问题即可开始。"}
            </p>
          </div>
        )}
        {messages.map((m) => (
          <div
            key={m.id}
            className={`max-w-[92%] md:max-w-xl rounded-2xl px-3.5 py-2.5 text-sm whitespace-pre-wrap leading-relaxed ${
              m.role === "user"
                ? "ml-auto bg-stone-900 text-stone-50"
                : `mr-auto bg-white border border-stone-200 text-stone-800 border-l-4 ${characterAccent(m.characterId)}`
            }`}
          >
            {m.role === "assistant" && m.characterName && (
              <p className="text-xs font-medium text-stone-600 mb-1.5">
                {m.characterName}
                {m.era ? ` · ${m.era}` : ""}
                {m.round ? ` · 第${m.round}轮` : ""}
              </p>
            )}
            {m.content}
            {m.streaming && (
              <span className="inline-block w-1.5 h-4 ml-0.5 bg-stone-400 animate-pulse align-middle" />
            )}
            {m.meta && (
              <p
                className={`text-[10px] mt-2 ${m.meta === "error" ? "text-red-600" : "opacity-60"}`}
              >
                {m.meta === "error" ? "错误" : m.meta}
              </p>
            )}
          </div>
        ))}
        {loading && (
          <p className="text-xs text-stone-500 animate-pulse px-1">
            {loadingHint || "人物思考中…"}
          </p>
        )}
      </div>

      <footer className="shrink-0 p-3 border-t border-stone-200 bg-white">
        {providerHint && (
          <p className="text-xs text-amber-700 mb-2 leading-relaxed">{providerHint}</p>
        )}
        <div className="flex gap-2 items-end">
          <textarea
            value={input}
            onChange={(e) => onInputChange(e.target.value)}
            rows={2}
            placeholder={
              mode === "group"
                ? "输入追问，或从热榜选择议题…"
                : "输入问题，或从热榜选择议题…"
            }
            className="flex-1 resize-none rounded-xl border border-stone-300 px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-amber-600/25"
            onKeyDown={(e) => {
              if (e.key === "Enter" && !e.shiftKey) {
                e.preventDefault();
                if (!loading && input.trim() && canSend) onSend();
              }
            }}
          />
          <button
            type="button"
            disabled={loading || !input.trim() || !canSend}
            title={!canSend ? "请先在设置中选择人物" : undefined}
            onClick={onSend}
            className="shrink-0 rounded-xl bg-stone-900 text-white px-5 text-sm font-medium disabled:opacity-40 min-h-[44px] hover:bg-stone-800 transition"
          >
            发送
          </button>
        </div>
      </footer>
    </section>
  );
}
