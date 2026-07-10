export type SSEEvent = {
  event: string;
  data: Record<string, unknown>;
};

export async function consumeSSE(
  url: string,
  body: unknown,
  onEvent: (ev: SSEEvent) => void,
): Promise<void> {
  const res = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json", Accept: "text/event-stream" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error((err as { error?: string }).error || `请求失败 (${res.status})`);
  }
  if (!res.body) {
    throw new Error("无响应流");
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });
    buffer = parseSSEBuffer(buffer, onEvent);
  }
  parseSSEBuffer(buffer + "\n\n", onEvent);
}

function parseSSEBuffer(buffer: string, onEvent: (ev: SSEEvent) => void): string {
  const parts = buffer.split("\n\n");
  const rest = parts.pop() ?? "";
  for (const block of parts) {
    const lines = block.split("\n");
    let event = "message";
    let data = "";
    for (const line of lines) {
      if (line.startsWith("event:")) {
        event = line.slice(6).trim();
      } else if (line.startsWith("data:")) {
        data = line.slice(5).trim();
      }
    }
    if (!data) continue;
    try {
      onEvent({ event, data: JSON.parse(data) as Record<string, unknown> });
    } catch {
      onEvent({ event, data: { raw: data } });
    }
  }
  return rest;
}
