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
  portrait?: string;
  tags?: string[];
  keyWorks?: string[];
  fitTopics?: string[];
  intro?: string;
}

export interface ScenarioItem {
  id: string;
  title: string;
  hook: string;
  mode: "single" | "group";
  characterIds: string[];
  hotItem: HotItem;
  sampleQuestion: string;
  expectedAngle?: string;
  demoId?: string;
}

export interface DemoMessage {
  role: "user" | "assistant";
  characterId?: string;
  characterName?: string;
  era?: string;
  content: string;
  round?: number;
}

export interface DemoConversation {
  id: string;
  scenarioId?: string;
  sourceTitle: string;
  mode: "single" | "group";
  characterIds?: string[];
  messages: DemoMessage[];
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

export interface Citation {
  title: string;
  source?: string;
  excerpt: string;
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
  citations?: Citation[];
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

export function characterBg(id?: string) {
  const map: Record<string, string> = {
    luxun: "bg-stone-800",
    sushi: "bg-teal-700",
    lihongzhang: "bg-amber-800",
    libai: "bg-indigo-700",
    zhugeliang: "bg-slate-700",
    wangyangming: "bg-emerald-800",
    zhuangzi: "bg-cyan-700",
    wuzetian: "bg-rose-800",
  };
  if (!id) return "bg-stone-500";
  return map[id] ?? "bg-stone-600";
}
