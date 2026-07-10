"use client";

import type { ReactNode } from "react";
import type { ChatMode } from "./ModeToggle";

interface CollapsibleSetupProps {
  collapsed: boolean;
  onToggle: () => void;
  chatMode: ChatMode;
  sourceTitle?: string;
  characterName?: string;
  groupNames?: string;
  showCollapse?: boolean;
  children: ReactNode;
}

export function CollapsibleSetup({
  collapsed,
  onToggle,
  chatMode,
  sourceTitle,
  characterName,
  groupNames,
  showCollapse = false,
  children,
}: CollapsibleSetupProps) {
  const participants =
    chatMode === "group"
      ? groupNames || "未选成员"
      : characterName || "未选人物";

  if (collapsed) {
    return (
      <div className="shrink-0 border-b border-stone-200 bg-[#f8f6f2]">
        <div className="max-w-5xl mx-auto px-4 py-2 flex items-center gap-3">
          <button
            type="button"
            onClick={onToggle}
            className="shrink-0 text-xs px-2.5 py-1 rounded-md border border-stone-300 bg-white hover:bg-stone-50 text-stone-700"
            aria-expanded={false}
          >
            展开
          </button>
          <p className="min-w-0 flex-1 text-xs text-stone-600 truncate">
            {sourceTitle ? (
              <>
                <span className="text-stone-500">议题</span> {sourceTitle}
                <span className="mx-1.5 text-stone-300">·</span>
              </>
            ) : null}
            <span className="text-stone-500">
              {chatMode === "group" ? "成员" : "人物"}
            </span>{" "}
            {participants}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="shrink-0 border-b border-stone-200 bg-gradient-to-b from-white to-[#f8f6f2]">
      {showCollapse ? (
        <div className="border-b border-stone-200/60">
          <div className="max-w-5xl mx-auto px-4 py-1.5 flex items-center justify-end">
            <button
              type="button"
              onClick={onToggle}
              className="text-xs px-2.5 py-1 rounded-md border border-stone-300 bg-white hover:bg-stone-50 text-stone-600"
              aria-expanded={true}
            >
              收起
            </button>
          </div>
        </div>
      ) : null}
      {children}
    </div>
  );
}
