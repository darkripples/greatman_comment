export type ChatMode = "single" | "group";

export const MIN_GROUP_MEMBERS = 2;
export const MAX_GROUP_MEMBERS = 5;

export interface HotItem {
  title: string;
  url: string;
  excerpt?: string;
  detail_text?: string;
  thumbnail?: string;
  is_mock?: boolean;
}

export interface SearchItem {
  title: string;
  url: string;
  excerpt?: string;
  type?: string;
  is_mock?: boolean;
}

export interface CharacterItem {
  id: string;
  name: string;
  era: string;
  summary: string;
}

export interface ProviderItem {
  id: string;
  name: string;
  available: boolean;
  model: string;
  default: boolean;
}

export interface ChatResult {
  conversationId: string;
  content: string;
  provider: string;
  model: string;
}

export interface GroupTurn {
  characterId: string;
  name: string;
  era: string;
  content: string;
  round: number;
}

export interface GroupDiscussResult {
  conversationId: string;
  turns: GroupTurn[];
  provider: string;
  model: string;
}

export interface ChatMessage {
  id: string;
  role: "user" | "assistant";
  content: string;
  characterId?: string;
  characterName?: string;
  era?: string;
  round?: number;
  meta?: string;
  streaming?: boolean;
}

export interface ConversationSummary {
  id: string;
  mode: string;
  sourceTitle?: string;
  hotUrl?: string;
  characterIds?: string[];
  provider?: string;
  createdAt: number;
  updatedAt: number;
}

const CHARACTER_COLORS: Record<string, string> = {
  luxun: "border-l-stone-800",
  sushi: "border-l-teal-700",
  lihongzhang: "border-l-amber-800",
  libai: "border-l-indigo-700",
  zhugeliang: "border-l-slate-700",
  wangyangming: "border-l-emerald-800",
  zhuangzi: "border-l-cyan-700",
  wuzetian: "border-l-rose-800",
};

export function characterAccent(id?: string) {
  if (!id) return "border-l-stone-400";
  return CHARACTER_COLORS[id] ?? "border-l-stone-500";
}
