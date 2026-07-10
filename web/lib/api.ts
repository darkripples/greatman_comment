import type { AppSettings, AppSettingsResponse } from "./settings";
import { defaultSettings, getApiBase } from "./settings";
import { consumeSSE } from "./sse";
import type {
  CharacterItem,
  ChatMessage,
  ChatResult,
  ConversationSummary,
  GroupDiscussResult,
  GroupTurn,
  HotItem,
  ProviderItem,
  SearchItem,
} from "./types";

async function request<T>(
  settings: AppSettings,
  path: string,
  init?: RequestInit,
): Promise<T> {
  const base = getApiBase(settings);
  const url = `${base}${path}`;
  const res = await fetch(url, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers || {}),
    },
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data.error || `请求失败 (${res.status})`);
  }
  return data as T;
}

function mapHistory(history: ChatMessage[]) {
  return history.slice(-12).map((m) => ({
    role: m.role,
    characterId: m.characterId,
    characterName: m.characterName,
    content: m.content,
    round: m.round,
  }));
}

export async function fetchAppSettings(settings: AppSettings = defaultSettings) {
  return request<AppSettingsResponse>(settings, "/api/settings");
}

export async function updateAppSettings(
  settings: AppSettings,
  patch: Partial<AppSettings>,
) {
  return request<AppSettingsResponse>(settings, "/api/settings", {
    method: "PUT",
    body: JSON.stringify(patch),
  });
}

export async function fetchHotList(settings: AppSettings, limit = 20) {
  const data = await request<{ items: HotItem[] }>(
    settings,
    `/api/hot-list?limit=${limit}`,
  );
  return data.items ?? [];
}

export async function fetchSearch(settings: AppSettings, q: string, count = 10) {
  const data = await request<{
    items: SearchItem[];
    cached?: boolean;
    stale?: boolean;
  }>(
    settings,
    `/api/search?q=${encodeURIComponent(q)}&count=${count}`,
  );
  return data.items ?? [];
}

export async function fetchCharacters(settings: AppSettings) {
  const data = await request<{ items: CharacterItem[] }>(
    settings,
    "/api/characters",
  );
  return data.items ?? [];
}

export async function fetchProviders(settings: AppSettings) {
  const data = await request<{ items: ProviderItem[] }>(
    settings,
    "/api/providers",
  );
  return data.items ?? [];
}

export async function sendChat(
  settings: AppSettings,
  body: {
    conversationId?: string;
    characterId: string;
    question: string;
    sourceTitle?: string;
    sourceExcerpt?: string;
    sourceDetail?: string;
    hotUrl?: string;
    provider: string;
    round?: number;
    history?: ChatMessage[];
  },
) {
  const history = body.history ? mapHistory(body.history) : undefined;
  return request<ChatResult>(settings, "/api/chat", {
    method: "POST",
    body: JSON.stringify({ ...body, history }),
  });
}

export async function sendChatStream(
  settings: AppSettings,
  body: {
    conversationId?: string;
    characterId: string;
    question: string;
    sourceTitle?: string;
    sourceExcerpt?: string;
    sourceDetail?: string;
    hotUrl?: string;
    provider: string;
    round?: number;
    history?: ChatMessage[];
  },
  onEvent: Parameters<typeof consumeSSE>[2],
) {
  const base = getApiBase(settings);
  const history = body.history ? mapHistory(body.history) : undefined;
  await consumeSSE(`${base}/api/chat/stream`, { ...body, history }, onEvent);
}

export async function sendGroupDiscuss(
  settings: AppSettings,
  body: {
    conversationId?: string;
    characterIds: string[];
    question: string;
    sourceTitle?: string;
    sourceExcerpt?: string;
    sourceDetail?: string;
    hotUrl?: string;
    provider: string;
    history: ChatMessage[];
    round: number;
    speakerIndex?: number;
    priorTurnsInRound?: GroupTurn[];
  },
) {
  const history = mapHistory(body.history);
  return request<GroupDiscussResult>(settings, "/api/group-discuss", {
    method: "POST",
    body: JSON.stringify({ ...body, history }),
  });
}

export async function sendGroupDiscussStream(
  settings: AppSettings,
  body: {
    conversationId?: string;
    characterIds: string[];
    question: string;
    sourceTitle?: string;
    sourceExcerpt?: string;
    sourceDetail?: string;
    hotUrl?: string;
    provider: string;
    history: ChatMessage[];
    round: number;
  },
  onEvent: Parameters<typeof consumeSSE>[2],
) {
  const base = getApiBase(settings);
  const history = mapHistory(body.history);
  await consumeSSE(`${base}/api/group-discuss/stream`, { ...body, history }, onEvent);
}

export async function fetchConversations(settings: AppSettings, limit = 20) {
  const data = await request<{ items: ConversationSummary[] }>(
    settings,
    `/api/conversations?limit=${limit}`,
  );
  return data.items ?? [];
}

export async function fetchConversation(settings: AppSettings, id: string) {
  return request<{
    conversation: ConversationSummary;
    messages: Array<{
      id: number;
      role: string;
      characterId?: string;
      characterName?: string;
      era?: string;
      round: number;
      content: string;
      provider?: string;
      model?: string;
    }>;
  }>(settings, `/api/conversations/${id}`);
}

export async function checkHealth(settings: AppSettings) {
  return request<{ status: string }>(settings, "/api/health");
}
