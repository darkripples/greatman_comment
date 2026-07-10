"use client";

export type ChatMode = "single" | "group";

const MODE_KEY = "renwen-chat-mode";

export function loadChatMode(): ChatMode {
  if (typeof window === "undefined") return "single";
  const v = localStorage.getItem(MODE_KEY);
  return v === "group" ? "group" : "single";
}

export function saveChatMode(mode: ChatMode) {
  localStorage.setItem(MODE_KEY, mode);
}

interface ModeToggleProps {
  mode: ChatMode;
  onChange: (mode: ChatMode) => void;
  compact?: boolean;
}

export function ModeToggle({ mode, onChange, compact }: ModeToggleProps) {
  if (compact) {
    return (
      <div className="inline-flex rounded-lg border border-stone-300 p-0.5 bg-stone-100">
        <button
          type="button"
          onClick={() => onChange("single")}
          className={`px-2.5 py-1.5 text-xs rounded-md transition ${
            mode === "single"
              ? "bg-white text-stone-900 shadow-sm"
              : "text-stone-600 hover:text-stone-900"
          }`}
        >
          单人
        </button>
        <button
          type="button"
          onClick={() => onChange("group")}
          className={`px-2.5 py-1.5 text-xs rounded-md transition ${
            mode === "group"
              ? "bg-white text-stone-900 shadow-sm"
              : "text-stone-600 hover:text-stone-900"
          }`}
        >
          群聊
        </button>
      </div>
    );
  }

  return (
    <section className="space-y-2">
      <h2 className="text-sm font-semibold text-stone-800">对话模式</h2>
      <div className="grid grid-cols-2 gap-2">
        <button
          type="button"
          onClick={() => onChange("single")}
          className={`rounded-lg border px-3 py-2 text-sm ${
            mode === "single"
              ? "border-stone-800 bg-stone-900 text-white"
              : "border-stone-200 bg-white"
          }`}
        >
          单人对话
        </button>
        <button
          type="button"
          onClick={() => onChange("group")}
          className={`rounded-lg border px-3 py-2 text-sm ${
            mode === "group"
              ? "border-stone-800 bg-stone-900 text-white"
              : "border-stone-200 bg-white"
          }`}
        >
          群聊讨论
        </button>
      </div>
    </section>
  );
}
