"use client";

import { useCallback, useEffect, useState } from "react";
import {
  checkHealth,
  fetchAppSettings,
  fetchCharacters,
  fetchConversation,
  fetchConversations,
  fetchHotList,
  fetchProviders,
  sendChatStream,
  sendGroupDiscussStream,
  updateAppSettings,
} from "@/lib/api";
import {
  defaultSettings,
  type AppSettings,
  type ChatModeId,
} from "@/lib/settings";
import type {
  ChatMessage,
  CharacterItem,
  ConversationSummary,
  HotItem,
  ProviderItem,
} from "@/lib/types";
import { MAX_GROUP_MEMBERS } from "@/lib/types";
import { AppHeader } from "@/components/AppHeader";
import { CharacterBar } from "@/components/CharacterBar";
import { ChatWindow } from "@/components/ChatWindow";
import { CollapsibleSetup } from "@/components/CollapsibleSetup";
import { ConversationHistoryPanel } from "@/components/ConversationHistoryPanel";
import { HotListPanel } from "@/components/HotListPanel";
import { type ChatMode } from "@/components/ModeToggle";
import { SettingsDrawer } from "@/components/SettingsDrawer";
import { SideDrawer } from "@/components/SideDrawer";

function uid() {
  return Math.random().toString(36).slice(2);
}

const DEFAULT_GROUP_IDS = ["luxun", "sushi", "lihongzhang"];

export default function HomePage() {
  const [settings, setSettings] = useState<AppSettings>(defaultSettings);
  const [hydrated, setHydrated] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [historyOpen, setHistoryOpen] = useState(false);

  const [hotItems, setHotItems] = useState<HotItem[]>([]);
  const [characters, setCharacters] = useState<CharacterItem[]>([]);
  const [providers, setProviders] = useState<ProviderItem[]>([]);

  const [hotLoading, setHotLoading] = useState(false);
  const [hotError, setHotError] = useState<string>();
  const [apiStatus, setApiStatus] = useState("");
  const [hasZhihuKey, setHasZhihuKey] = useState(false);
  const [hasDeepSeekKey, setHasDeepSeekKey] = useState(false);

  const [selectedHot, setSelectedHot] = useState<HotItem>();
  const [selectedCharacter, setSelectedCharacter] = useState<CharacterItem>();
  const [selectedGroupIds, setSelectedGroupIds] = useState<string[]>(DEFAULT_GROUP_IDS);

  const [conversationId, setConversationId] = useState<string>();
  const [round, setRound] = useState(1);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState("");
  const [chatLoading, setChatLoading] = useState(false);
  const [loadingHint, setLoadingHint] = useState<string>();

  const [historyItems, setHistoryItems] = useState<ConversationSummary[]>([]);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [historyError, setHistoryError] = useState<string>();
  const [setupExpanded, setSetupExpanded] = useState(false);

  const chatMode: ChatMode =
    settings.chatMode === "group" ? "group" : "single";

  const applySettingsResponse = useCallback(
    (res: AppSettings & { hasZhihuKey?: boolean; hasDeepSeekKey?: boolean }) => {
      setSettings({
        apiEnvironment: res.apiEnvironment as AppSettings["apiEnvironment"],
        localApiMode: res.localApiMode as AppSettings["localApiMode"],
        prodApiBase: res.prodApiBase,
        devApiBase: res.devApiBase,
        llmProvider: res.llmProvider as AppSettings["llmProvider"],
        deepseekModel: res.deepseekModel,
        zhidaModel: res.zhidaModel,
        deepseekApiBase: res.deepseekApiBase,
        zhihuApiBase: res.zhihuApiBase,
        zhihuMock: res.zhihuMock,
        hotListCacheTtl: res.hotListCacheTtl,
        hotListMinInterval: res.hotListMinInterval,
        searchCacheTtl: res.searchCacheTtl,
        searchMinInterval: res.searchMinInterval,
        chatMode: (res.chatMode === "group" ? "group" : "single") as ChatModeId,
      });
      if (res.hasZhihuKey != null) setHasZhihuKey(res.hasZhihuKey);
      if (res.hasDeepSeekKey != null) setHasDeepSeekKey(res.hasDeepSeekKey);
    },
    [],
  );

  const persistSettings = useCallback(
    async (next: AppSettings, prev: AppSettings) => {
      setSettings(next);
      try {
        const res = await updateAppSettings(prev, next);
        applySettingsResponse(res);
      } catch (e) {
        setApiStatus(e instanceof Error ? e.message : "保存设置失败");
      }
    },
    [applySettingsResponse],
  );

  const handleChatModeChange = (mode: ChatMode) => {
    void persistSettings({ ...settings, chatMode: mode }, settings);
  };

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const res = await fetchAppSettings(defaultSettings);
        if (!cancelled) {
          applySettingsResponse(res);
          setHydrated(true);
        }
      } catch (e) {
        if (!cancelled) {
          setApiStatus(e instanceof Error ? e.message : "加载设置失败");
          setHydrated(true);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [applySettingsResponse]);

  const reloadHistory = useCallback(async (s: AppSettings) => {
    setHistoryLoading(true);
    setHistoryError(undefined);
    try {
      const items = await fetchConversations(s, 30);
      setHistoryItems(items);
    } catch (e) {
      setHistoryError(e instanceof Error ? e.message : "加载历史失败");
    } finally {
      setHistoryLoading(false);
    }
  }, []);

  const reloadMeta = useCallback(async (s: AppSettings) => {
    setHotLoading(true);
    setHotError(undefined);
    try {
      const [chars, prov] = await Promise.all([
        fetchCharacters(s),
        fetchProviders(s),
      ]);
      setCharacters(chars);
      setProviders(prov);
      setSelectedCharacter((prev) => prev ?? chars[0]);
      const current = prov.find((p) => p.id === s.llmProvider);
      if (current && !current.available) {
        const fallback =
          prov.find((p) => p.default && p.available)?.id ||
          prov.find((p) => p.available)?.id;
        if (fallback) {
          setSettings((prev) => ({
            ...prev,
            llmProvider: fallback as AppSettings["llmProvider"],
          }));
        }
      }
    } catch (e) {
      setApiStatus(e instanceof Error ? e.message : "加载人物/模型失败");
    }

    try {
      const hotItems = await fetchHotList(s);
      setHotItems(hotItems);
    } catch (e) {
      setHotError(e instanceof Error ? e.message : "加载热榜失败");
    } finally {
      setHotLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!hydrated) return;
    reloadMeta(settings);
    reloadHistory(settings);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [hydrated, settings.apiEnvironment, settings.localApiMode, settings.prodApiBase, settings.devApiBase, settings.zhihuMock, reloadMeta, reloadHistory]);

  const resetDiscussion = () => {
    setConversationId(undefined);
    setRound(1);
    setMessages([]);
    setInput("");
    setSetupExpanded(false);
  };

  const discussionActive = messages.length > 0;
  const setupCollapsed = discussionActive && !setupExpanded;

  const loadConversation = async (item: ConversationSummary) => {
    setChatLoading(true);
    setLoadingHint("加载历史对话…");
    setHistoryOpen(false);
    try {
      const detail = await fetchConversation(settings, item.id);
      const conv = detail.conversation;
      setConversationId(conv.id);
      if (conv.mode === "group" || conv.mode === "single") {
        void persistSettings(
          { ...settings, chatMode: conv.mode === "group" ? "group" : "single" },
          settings,
        );
      }
      if (conv.sourceTitle) {
        setSelectedHot({ title: conv.sourceTitle, url: conv.hotUrl || "" });
        setInput(conv.sourceTitle);
      }
      if (conv.mode === "group" && conv.characterIds?.length) {
        setSelectedGroupIds(conv.characterIds);
      } else if (conv.characterIds?.[0]) {
        const ch = characters.find((c) => c.id === conv.characterIds![0]);
        if (ch) setSelectedCharacter(ch);
      }
      const loaded: ChatMessage[] = detail.messages.map((m) => ({
        id: String(m.id),
        role: m.role as "user" | "assistant",
        content: m.content,
        characterId: m.characterId,
        characterName: m.characterName,
        era: m.era,
        round: m.round,
        meta: m.provider && m.model ? `${m.provider} · ${m.model}` : undefined,
      }));
      setMessages(loaded);
      const maxRound = loaded.reduce((max, m) => Math.max(max, m.round || 1), 1);
      setRound(maxRound + 1);
      setSetupExpanded(false);
    } catch (e) {
      setHistoryError(e instanceof Error ? e.message : "加载对话失败");
    } finally {
      setChatLoading(false);
      setLoadingHint(undefined);
    }
  };

  const handleHotSelect = (item: HotItem) => {
    setSelectedHot(item);
    setInput(item.title);
    resetDiscussion();
  };

  const canSendGroup = selectedGroupIds.length >= 2;
  const canSend = chatMode === "single" ? !!selectedCharacter : canSendGroup;
  const currentProvider = providers.find((p) => p.id === settings.llmProvider);
  const providerReady = currentProvider?.available ?? true;
  const groupCharacters = characters.filter((c) => selectedGroupIds.includes(c.id));
  const groupNames = groupCharacters.map((c) => c.name).join("、");

  const handleSend = async () => {
    const question = input.trim();
    if (!question || !canSend || !providerReady) return;

    const userMsg: ChatMessage = {
      id: uid(),
      role: "user",
      content: question,
      round,
    };
    setMessages((prev) => [...prev, userMsg]);
    setInput("");
    setChatLoading(true);

    const sourcePayload = {
      sourceTitle: selectedHot?.title,
      sourceExcerpt: selectedHot?.excerpt,
      sourceDetail: selectedHot?.detail_text,
      hotUrl: selectedHot?.url,
    };
    const history = [...messages, userMsg];

    try {
      if (chatMode === "single" && selectedCharacter) {
        const assistantId = uid();
        setLoadingHint(`${selectedCharacter.name} 思考中…`);
        setMessages((prev) => [
          ...prev,
          {
            id: assistantId,
            role: "assistant",
            characterId: selectedCharacter.id,
            characterName: selectedCharacter.name,
            era: selectedCharacter.era,
            content: "",
            round,
            streaming: true,
          },
        ]);

        let cid = conversationId;
        let meta = "";

        await sendChatStream(
          settings,
          {
            conversationId,
            characterId: selectedCharacter.id,
            question,
            ...sourcePayload,
            provider: settings.llmProvider,
            round,
            history,
          },
          ({ event, data }) => {
            if (event === "meta" && typeof data.conversationId === "string") {
              cid = data.conversationId;
            }
            if (event === "delta" && typeof data.content === "string") {
              const delta = data.content;
              setMessages((prev) =>
                prev.map((m) =>
                  m.id === assistantId
                    ? { ...m, content: m.content + delta }
                    : m,
                ),
              );
            }
            if (event === "done") {
              if (typeof data.conversationId === "string") cid = data.conversationId;
              const provider = String(data.provider || "");
              const model = String(data.model || "");
              meta = provider && model ? `${provider} · ${model}` : provider;
              const content = typeof data.content === "string" ? data.content : undefined;
              if (content) {
                setMessages((prev) =>
                  prev.map((m) =>
                    m.id === assistantId
                      ? { ...m, content, streaming: false, meta }
                      : m,
                  ),
                );
              }
            }
            if (event === "error") {
              throw new Error(String(data.message || "对话失败"));
            }
          },
        );

        setConversationId(cid);
        setRound((r) => r + 1);
        setMessages((prev) =>
          prev.map((m) =>
            m.id === assistantId ? { ...m, streaming: false, meta: meta || m.meta } : m,
          ),
        );
      } else {
        setLoadingHint("群聊讨论中…");
        let cid = conversationId;
        let meta = "";

        await sendGroupDiscussStream(
          settings,
          {
            conversationId,
            characterIds: selectedGroupIds,
            question,
            ...sourcePayload,
            provider: settings.llmProvider,
            history,
            round,
          },
          ({ event, data }) => {
            if (event === "meta" && typeof data.conversationId === "string") {
              cid = data.conversationId;
            }
            if (event === "turn_start") {
              const charId = String(data.characterId || "");
              setLoadingHint(`${String(data.name || charId)} 发言中…`);
              setMessages((prev) => [
                ...prev,
                {
                  id: uid(),
                  role: "assistant",
                  characterId: charId,
                  characterName: String(data.name || ""),
                  era: String(data.era || ""),
                  round: Number(data.round) || round,
                  content: "",
                  streaming: true,
                },
              ]);
            }
            if (event === "turn_done") {
              const charId = String(data.characterId || "");
              const content = String(data.content || "");
              setMessages((prev) => {
                const idx = [...prev].reverse().findIndex(
                  (m) => m.characterId === charId && m.streaming,
                );
                if (idx < 0) return prev;
                const realIdx = prev.length - 1 - idx;
                return prev.map((m, i) =>
                  i === realIdx
                    ? { ...m, content, streaming: false }
                    : m,
                );
              });
            }
            if (event === "done") {
              if (typeof data.conversationId === "string") cid = data.conversationId;
              const provider = String(data.provider || "");
              const model = String(data.model || "");
              meta = provider && model ? `${provider} · ${model}` : provider;
            }
            if (event === "error") {
              throw new Error(String(data.message || "群聊失败"));
            }
          },
        );

        setConversationId(cid);
        setRound((r) => r + 1);
        if (meta) {
          setMessages((prev) =>
            prev.map((m) =>
              m.role === "assistant" && m.round === round && !m.meta
                ? { ...m, meta, streaming: false }
                : m,
            ),
          );
        }
      }
      void reloadHistory(settings);
    } catch (e) {
      setMessages((prev) => [
        ...prev.filter((m) => !m.streaming),
        {
          id: uid(),
          role: "assistant",
          content: e instanceof Error ? e.message : "对话失败",
          meta: "error",
        },
      ]);
    } finally {
      setChatLoading(false);
      setLoadingHint(undefined);
    }
  };

  const toggleGroupId = (id: string) => {
    setSelectedGroupIds((prev) => {
      if (prev.includes(id)) {
        return prev.filter((x) => x !== id);
      }
      if (prev.length >= MAX_GROUP_MEMBERS) return prev;
      return [...prev, id];
    });
  };

  const testConnection = async () => {
    try {
      const res = await checkHealth(settings);
      setApiStatus(`连接成功：${res.status}`);
    } catch (e) {
      setApiStatus(e instanceof Error ? e.message : "连接失败");
    }
  };

  return (
    <div className="h-[100dvh] flex flex-col bg-[#f3efe8] text-stone-900 overflow-hidden">
      <AppHeader
        chatMode={chatMode}
        onChatModeChange={handleChatModeChange}
        onOpenSettings={() => setSettingsOpen(true)}
        onOpenHistory={() => setHistoryOpen(true)}
        historyCount={historyItems.length}
      />

      <CollapsibleSetup
        collapsed={setupCollapsed}
        onToggle={() => setSetupExpanded((v) => !v)}
        chatMode={chatMode}
        sourceTitle={selectedHot?.title}
        characterName={selectedCharacter?.name}
        groupNames={groupNames}
        showCollapse={discussionActive}
      >
        <HotListPanel
          settings={settings}
          items={hotItems}
          selectedTitle={selectedHot?.title}
          loading={hotLoading}
          error={hotError}
          onSelect={handleHotSelect}
          onRefresh={() => reloadMeta(settings)}
        />

        <CharacterBar
          mode={chatMode}
          items={characters}
          selectedId={selectedCharacter?.id}
          selectedGroupIds={selectedGroupIds}
          onSelectSingle={setSelectedCharacter}
          onToggleGroup={toggleGroupId}
        />
      </CollapsibleSetup>

      <main className="flex-1 min-h-0 max-w-5xl w-full mx-auto px-4 py-3 md:py-4">
        <ChatWindow
          mode={chatMode}
          messages={messages}
          input={input}
          loading={chatLoading}
          loadingHint={loadingHint}
          characterName={selectedCharacter?.name}
          groupCharacters={groupCharacters}
          sourceTitle={selectedHot?.title}
          canSend={canSend && providerReady}
          providerHint={
            !providerReady
              ? `当前 LLM（${currentProvider?.name || settings.llmProvider}）未配置 Key，请在设置中切换提供方，或在系统环境变量中配置 DEEPSEEK_API_KEY / ZHIHU_API_KEY 后重启后端`
              : undefined
          }
          onInputChange={setInput}
          onSend={handleSend}
          onNewDiscussion={resetDiscussion}
        />
      </main>

      <SettingsDrawer
        open={settingsOpen}
        onClose={() => setSettingsOpen(false)}
        settings={settings}
        providers={providers}
        apiStatus={apiStatus}
        hasZhihuKey={hasZhihuKey}
        hasDeepSeekKey={hasDeepSeekKey}
        onSettingsChange={(next) => void persistSettings(next, settings)}
        onTestConnection={testConnection}
      />

      <SideDrawer open={historyOpen} title="历史记录" onClose={() => setHistoryOpen(false)}>
        <ConversationHistoryPanel
          items={historyItems}
          activeId={conversationId}
          loading={historyLoading}
          error={historyError}
          embedded
          onSelect={loadConversation}
          onRefresh={() => reloadHistory(settings)}
        />
      </SideDrawer>
    </div>
  );
}
