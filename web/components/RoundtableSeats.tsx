import type { CharacterItem, ChatMessage } from "@/lib/types";
import { characterBg } from "@/lib/types";

export function RoundtableSeats({ characters, messages }: { characters: CharacterItem[]; messages: ChatMessage[] }) {
  const active = [...messages].reverse().find((message) => message.streaming)?.characterId;
  const spoken = new Set(messages.filter((message) => message.role === "assistant" && !message.streaming).map((message) => message.characterId));
  return (
    <div className="flex gap-2 overflow-x-auto py-1" aria-label="圆桌发言席位">
      {characters.map((character) => {
        const isActive = character.id === active;
        const hasSpoken = spoken.has(character.id);
        return <div key={character.id} className={`shrink-0 flex items-center gap-1.5 rounded-full border px-2 py-1 text-[11px] ${isActive ? "border-amber-600 bg-amber-50 text-amber-900" : hasSpoken ? "border-stone-300 bg-white text-stone-700" : "border-stone-200 bg-stone-50 text-stone-400"}`}>
          <span className={`flex h-4 w-4 items-center justify-center rounded-full text-[9px] text-white ${characterBg(character.id)}`}>{character.name.slice(0, 1)}</span>
          <span>{character.name}</span>
          {isActive && <span className="animate-pulse">发言中</span>}
        </div>;
      })}
    </div>
  );
}
