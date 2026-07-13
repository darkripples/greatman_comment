"use client";

import type { ChatMessage, CharacterItem } from "@/lib/types";
import { MAX_GROUP_MEMBERS, characterAccent, characterBg } from "@/lib/types";
import type { ChatMode } from "./ModeToggle";
import { CitationBlock } from "./CitationBlock";

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

function RoundDivider({ round }: { round: number }) {
  return (
    <div className="flex items-center gap-3 py-2">
      <div className="flex-1 h-px bg-stone-200" />
      <span className="text-[10px] text-stone-500 font-serif shrink-0">
        第 {round} 轮圆桌
      </span>
      <div className="flex-1 h-px bg-stone-200" />
    </div>
  );
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

  let lastRound = 0;

  return (
    <section className="flex flex-col h-full min-h-0 rounded-2xl border border-stone-200 bg-white shadow-sm overflow-hidden">
      <header className="px-4 py-3 border-b border-stone-200 bg-stone-50/80 flex flex-wrap items-start justify-between gap-2">
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <p className="text-sm font-medium text-stone-800 font-serif">
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
              选择精选场景，或从热榜挑选议题开始
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
            <p className="text-base text-stone-700 font-serif">跨越时空的对话</p>
            <p className="text-xs text-stone-500 mt-2 max-w-md leading-relaxed">
              从上方「精选场景」一键开始，或点击「观看精选演示」无需 API Key 即可体验。
            </p>
            <p className="text-xs text-stone-400 mt-2 max-w-md leading-relaxed">
              {mode === "group"
                ? `群聊模式：选择 2–${MAX_GROUP_MEMBERS} 位人物，看不同价值观如何碰撞。`
                : "单人模式：选一位历史人物，以有限视角点评今日议题。"}
            </p>
          </div>
        )}
        {messages.map((m) => {
          const showRoundDivider =
            mode === "group" &&
            m.round != null &&
            m.round > 0 &&
            m.round !== lastRound &&
            (m.role === "user" || m.characterName);
          if (showRoundDivider && m.round != null) {
            lastRound = m.round;
          }
          return (
            <div key={m.id}>
              {showRoundDivider && m.round != null && (
                <RoundDivider round={m.round} />
              )}
              <div
                className={`max-w-[92%] md:max-w-xl rounded-2xl px-3.5 py-2.5 text-sm whitespace-pre-wrap leading-relaxed ${
                  m.role === "user"
                    ? "ml-auto bg-stone-900 text-stone-50"
                    : `mr-auto bg-white border border-stone-200 text-stone-800 border-l-4 ${characterAccent(m.characterId)}`
                }`}
              >
                {m.role === "assistant" && m.characterName && (
                  <div className="flex items-center gap-2 mb-1.5">
                    <span
                      className={`inline-flex w-6 h-6 rounded-full text-[10px] text-white items-center justify-center shrink-0 ${characterBg(m.characterId)}`}
                    >
                      {m.characterName.slice(0, 1)}
                    </span>
                    <p className="text-xs font-medium text-stone-600">
                      {m.characterName}
                      {m.era ? ` · ${m.era}` : ""}
                    </p>
                  </div>
                )}
                {m.content}
                {m.streaming && (
                  <span className="inline-block w-1.5 h-4 ml-0.5 bg-stone-400 animate-pulse align-middle" />
                )}
                {m.meta && (
                  <p
                    className={`text-[10px] mt-2 ${m.meta === "error" ? "text-red-600" : m.meta === "demo" ? "text-amber-700" : "opacity-60"}`}
                  >
                    {m.meta === "error" ? "错误" : m.meta === "demo" ? "精选演示回放" : m.meta}
                  </p>
                )}
                {m.role === "assistant" && (
                  <CitationBlock citations={m.citations} />
                )}
              </div>
            </div>
          );
        })}
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
                ? "输入追问，或从精选场景开始…"
                : "输入问题，或从精选场景开始…"
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
            title={!canSend ? "请先在上方选择人物" : undefined}
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
