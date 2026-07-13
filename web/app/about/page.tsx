import Link from "next/link";

export default function AboutPage() {
  return (
    <div className="min-h-screen bg-[#f3efe8] text-stone-900">
      <div className="max-w-3xl mx-auto px-4 py-10 space-y-8">
        <header>
          <Link href="/" className="text-xs text-stone-500 hover:text-stone-800">
            ← 返回对话
          </Link>
          <h1 className="text-2xl font-serif font-semibold mt-4">用 AI 重新看见人</h1>
          <p className="text-sm text-stone-600 mt-2">人文季 · 历史单元 · greatman_comment</p>
        </header>

        <section className="space-y-3 text-sm leading-relaxed text-stone-700">
          <h2 className="font-serif font-semibold text-stone-900">核心命题</h2>
          <p>
            把「今日之问」交给「昨日之人」——不是让历史人物全知穿越，而是在<strong>时代边界</strong>内，
            以各自的价值立场与思想资源，点评知乎热榜议题。
          </p>
        </section>

        <section className="space-y-3 text-sm leading-relaxed text-stone-700">
          <h2 className="font-serif font-semibold text-stone-900">方法论</h2>
          <ul className="list-disc pl-5 space-y-2">
            <li><strong>时代边界</strong>：人物不装全知，超出时代的问题在正文中自然点明局限。</li>
            <li><strong>史料优先</strong>：基于可考思想立场与语料片段回应，减少随意编造。</li>
            <li><strong>群聊碰撞</strong>：2–5 位人物在同一议题下串行发言，形成多视角圆桌。</li>
          </ul>
        </section>

        <section className="space-y-3 text-sm leading-relaxed text-stone-700">
          <h2 className="font-serif font-semibold text-stone-900">使用方式</h2>
          <ol className="list-decimal pl-5 space-y-2">
            <li>从「精选场景」一键填充议题、人物与首问，或点击「观看精选演示」无需 API Key 体验。</li>
            <li>也可从知乎热榜（Mock 模式可用本地 fixtures）选择议题。</li>
            <li>单人深聊或群聊圆桌，流式 SSE 实时输出。</li>
          </ol>
        </section>

        <section className="space-y-3 text-sm leading-relaxed text-stone-700">
          <h2 className="font-serif font-semibold text-stone-900">边界声明</h2>
          <p>
            本项目生成的是 AI「文本人格」，基于 Prompt 与有限语料模拟历史人物口吻，
            <strong>非史实复原，不可替代专业史学研究</strong>。欢迎批判性使用。
          </p>
        </section>

        <section className="space-y-3 text-sm leading-relaxed text-stone-700">
          <h2 className="font-serif font-semibold text-stone-900">技术栈</h2>
          <p>Go 后端 + Next.js 15 前端 · DeepSeek / 知乎直答 LLM · SQLite 持久化</p>
        </section>
      </div>
    </div>
  );
}
