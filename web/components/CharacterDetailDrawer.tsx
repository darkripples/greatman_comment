"use client";

import type { CharacterItem } from "@/lib/types";
import { characterAccent } from "@/lib/types";
import { SideDrawer } from "./SideDrawer";

interface CharacterDetailDrawerProps {
  open: boolean;
  character?: CharacterItem;
  onClose: () => void;
}

export function CharacterDetailDrawer({
  open,
  character,
  onClose,
}: CharacterDetailDrawerProps) {
  return (
    <SideDrawer open={open} title={character ? `${character.name} · 人物志` : "人物志"} onClose={onClose}>
      {!character ? (
        <p className="text-sm text-stone-500">未选择人物</p>
      ) : (
        <div className="space-y-4 text-sm text-stone-700">
          <div className={`rounded-xl border border-l-4 p-4 bg-stone-50 ${characterAccent(character.id)}`}>
            <div className="flex items-baseline justify-between gap-2">
              <h3 className="text-lg font-serif font-semibold text-stone-900">
                {character.name}
              </h3>
              <span className="text-xs text-stone-500">{character.era}</span>
            </div>
            <p className="text-xs text-stone-600 mt-2 leading-relaxed">
              {character.intro || character.summary}
            </p>
          </div>
          {character.tags && character.tags.length > 0 && (
            <div>
              <p className="text-xs font-medium text-stone-800 mb-2">风格标签</p>
              <div className="flex flex-wrap gap-1.5">
                {character.tags.map((t) => (
                  <span
                    key={t}
                    className="text-[11px] px-2 py-0.5 rounded-full bg-stone-200/80 text-stone-700"
                  >
                    {t}
                  </span>
                ))}
              </div>
            </div>
          )}
          {character.keyWorks && character.keyWorks.length > 0 && (
            <div>
              <p className="text-xs font-medium text-stone-800 mb-2">代表作品</p>
              <ul className="text-xs space-y-1 text-stone-600">
                {character.keyWorks.map((w) => (
                  <li key={w}>· {w}</li>
                ))}
              </ul>
            </div>
          )}
          {character.fitTopics && character.fitTopics.length > 0 && (
            <div>
              <p className="text-xs font-medium text-stone-800 mb-2">适合议题</p>
              <div className="flex flex-wrap gap-1.5">
                {character.fitTopics.map((t) => (
                  <span
                    key={t}
                    className="text-[11px] px-2 py-0.5 rounded-full border border-stone-300 text-stone-600"
                  >
                    {t}
                  </span>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </SideDrawer>
  );
}
