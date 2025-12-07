<script lang="ts">
  import { fade } from "svelte/transition";
  import CommandOutput from "./CommandOutput.svelte";
  import logo from "../assets/images/tf-logo.png";
  import AnsiToHtml from "ansi-to-html";
  import { marked } from "marked";
  import DOMPurify from "dompurify";
  import { BrowserOpenURL } from "../../wailsjs/runtime/runtime.js";

  // Step types
  const STEP_TYPE_TOOL = "tool";
  const STEP_TYPE_ANALYSIS = "analysis";
  const STEP_TYPE_QUESTION = "question";
  const STEP_TYPE_ANSWER = "answer";
  const STEP_TYPE_ERROR = "error";

  export let message: {
    role: string;
    content: string;
    timestamp: string;
    requestID?: string;
    steps?: Array<{
      type: string;
      progressText: string;
      content: string;
      output: string;
      error: string;
    }>;
    // Deprecated fields (backward compatibility)
    isCommand?: boolean;
    output?: string;
    error?: string;
  };

  const isUser = message.role === "user";
  let showSteps = false;

  // Helper function to check if step should be numbered
  function shouldNumberStep(step: { type?: string }): boolean {
    return (
      step.type !== STEP_TYPE_ANALYSIS &&
      step.type !== STEP_TYPE_QUESTION &&
      step.type !== STEP_TYPE_ANSWER
    );
  }

  // Calculate step numbers excluding analysis, question, and answer steps
  // Optimized: O(n) single-pass algorithm instead of O(n²)
  $: stepNumbers = message.steps
    ? (() => {
        let count = 0;
        return message.steps.map((step) => {
          if (!shouldNumberStep(step)) return null;
          return ++count;
        });
      })()
    : [];

  // Count of visible steps (excluding analysis, question, and answer steps)
  $: visibleStepCount = message.steps
    ? message.steps.filter(shouldNumberStep).length
    : 0;

  // ANSI to HTML converter
  const ansiConverter = new AnsiToHtml({
    fg: "#d4d4d4",
    bg: "#1e1e1e",
    newline: true,
    escapeXML: true,
  });

  // Convert ANSI codes to HTML
  function renderAnsi(text: string): string {
    if (!text) return "";
    return ansiConverter.toHtml(text);
  }

  // Render Markdown to HTML
  function renderMarkdown(text: string): string {
    if (!text) return "";

    // Configure marked options
    // breaks: true ensures that single newlines are converted to <br>,
    // which is better for "normal text" that isn't strictly markdown formatted.
    // gfm: true enables GitHub Flavored Markdown (tables, etc.)
    marked.setOptions({
      gfm: true,
      breaks: true,
    });

    // marked.parse returns a string or Promise<string>. In sync mode (default), it's string.
    const rawHtml = marked.parse(text) as string;
    return DOMPurify.sanitize(rawHtml);
  }

  // Handle link clicks to open in external browser
  function handleLinkClick(event: MouseEvent) {
    const target = event.target as HTMLElement;

    // Check if clicked element is a link or inside a link
    const link = target.closest("a");
    if (link && link.href) {
      event.preventDefault();
      BrowserOpenURL(link.href);
    }
  }
</script>

<div class="message-wrapper {isUser ? 'user' : 'agent'}" in:fade>
  <div class="avatar">
    {#if isUser}
      👤
    {:else}
      <img src={logo} alt="Agent" />
    {/if}
  </div>

  <div class="content-wrapper">
    <div class="bubble">
      {#if message.content}
        <div class="text markdown-body" on:click={handleLinkClick}>
          {@html renderMarkdown(message.content)}
        </div>
      {:else if message.steps && message.steps.length > 0}
        <div class="typing-indicator">
          <span></span>
          <span></span>
          <span></span>
        </div>
      {:else if message.role === "agent"}
        <div class="typing-indicator">
          <span></span>
          <span></span>
          <span></span>
        </div>
      {/if}
    </div>

    <!-- New: Steps display (Option 3 - Rich Message) -->
    {#if message.steps && message.steps.length > 0}
      <div class="steps-container">
        <button class="steps-toggle" on:click={() => (showSteps = !showSteps)}>
          <span class="toggle-icon">{showSteps ? "▼" : "▶"}</span>
          Show workflow ({visibleStepCount}
          {visibleStepCount === 1 ? "step" : "steps"})
        </button>

        {#if showSteps}
          <div class="steps" transition:fade>
            {#each message.steps as step, i}
              <div class="step">
                <div class="step-header">
                  {#if shouldNumberStep(step)}
                    <span class="step-number">{stepNumbers[i]}</span>
                  {/if}
                  <span class="step-type">
                    {step.progressText}
                  </span>
                </div>

                <div class="step-content">
                  <div class="step-command">{step.content}</div>

                  {#if step.output}
                    <div class="step-output">
                      <div class="output-label">
                        {step.type === "url_fetch"
                          ? "📄 Content:"
                          : "📤 Output:"}
                      </div>
                      <pre class="ansi-output">{@html renderAnsi(
                          step.output,
                        )}</pre>
                    </div>
                  {/if}

                  {#if step.error}
                    <div class="step-error">
                      <div class="error-label">❌ Error:</div>
                      <pre class="ansi-output">{@html renderAnsi(
                          step.error,
                        )}</pre>
                    </div>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    {/if}

    <!-- Backward compatibility: Old command display -->
    {#if message.isCommand && !message.steps}
      <CommandOutput output={message.output} error={message.error} />
    {/if}

    {#if message.content}
      <div class="timestamp">
        {new Date(message.timestamp).toLocaleTimeString()}
      </div>
    {/if}
  </div>
</div>

<style>
  .message-wrapper {
    display: flex;
    gap: 1rem;
    max-width: 85%;
    animation: slideUp 0.3s cubic-bezier(0.2, 0.8, 0.2, 1);
  }

  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translateY(10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .message-wrapper.user {
    margin-left: auto;
    flex-direction: row-reverse;
  }

  .avatar {
    width: 36px;
    height: 36px;
    border-radius: 12px; /* Slightly squircle */
    background: var(--bg-tertiary);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.25rem;
    flex-shrink: 0;
    overflow: hidden;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  }

  .avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .content-wrapper {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    min-width: 0;
  }

  .bubble {
    padding: 1rem 1.25rem;
    border-radius: 1.25rem;
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-word;
    overflow-wrap: anywhere;
    text-align: left;
    font-size: 0.95rem;
  }

  .user .bubble {
    background: var(--accent-gradient);
    color: white;
    border-bottom-right-radius: 0.25rem;
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.25);
  }

  .agent .bubble {
    background: var(--bg-secondary);
    color: var(--text-primary);
    border-top-left-radius: 0.25rem;
    border: 1px solid var(--border);
  }

  /* Premium typing indicator */
  .typing-indicator {
    display: flex;
    gap: 0.3rem;
    padding: 0.5rem 0.25rem;
  }

  .typing-indicator span {
    width: 6px;
    height: 6px;
    background: var(--text-secondary);
    border-radius: 50%;
    animation: flow 1.4s infinite ease-in-out both;
  }

  @keyframes flow {
    0%,
    80%,
    100% {
      opacity: 0.4;
      transform: translateY(0);
    }
    40% {
      opacity: 1;
      transform: translateY(-4px);
    }
  }

  .typing-indicator span:nth-child(1) {
    animation-delay: 0s;
  }
  .typing-indicator span:nth-child(2) {
    animation-delay: 0.15s;
  }
  .typing-indicator span:nth-child(3) {
    animation-delay: 0.3s;
  }

  /* Steps styling */
  .steps-container {
    margin-top: 0.75rem;
    border-top: 1px solid var(--border);
    padding-top: 0.75rem;
  }

  .steps-toggle {
    background: transparent;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    font-size: 0.875rem;
    padding: 0.25rem 0;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    transition: color 0.2s;
  }

  .steps-toggle:hover {
    color: var(--text-primary);
  }

  .toggle-icon {
    font-size: 0.75rem;
  }

  .steps {
    margin-top: 0.5rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .step {
    background: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    border-radius: 0.5rem;
    padding: 0.75rem;
  }

  .step-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .step-number {
    background: var(--accent);
    color: white;
    width: 1.5rem;
    height: 1.5rem;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.75rem;
  }

  .step-type {
    font-size: 0.875rem;
  }

  .step-content {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .step-command {
    font-family: "Courier New", monospace;
    background: var(--bg-primary);
    padding: 0.5rem;
    border-radius: 0.25rem;
    font-size: 0.875rem;
    color: var(--text-primary);
    white-space: pre-wrap;
    word-break: break-all;
    overflow-wrap: anywhere;
    text-align: left;
    overflow-x: auto;
  }

  .step-output,
  .step-error {
    margin-top: 0.25rem;
  }

  .output-label,
  .error-label {
    font-size: 0.75rem;
    font-weight: 600;
    margin-bottom: 0.25rem;
    color: var(--text-secondary);
  }

  .error-label {
    color: #ef4444;
  }

  .step-output pre,
  .step-error pre {
    background: var(--bg-primary);
    padding: 0.5rem;
    border-radius: 0.25rem;
    font-size: 0.75rem;
    overflow-x: auto;
    margin: 0;
    white-space: pre-wrap;
    word-break: break-all;
    overflow-wrap: anywhere;
    color: var(--text-secondary);
    text-align: left;
  }

  .step-error pre {
    color: #ef4444;
  }

  /* ANSI output styling */
  .ansi-output {
    font-family: "Courier New", Consolas, Monaco, monospace;
    line-height: 1.4;
  }

  /* Override ansi-to-html default styles to match our theme */
  .ansi-output :global(span) {
    font-family: inherit;
  }

  /* Markdown Styling */
  .markdown-body :global(p) {
    margin-bottom: 0.5rem;
  }

  .markdown-body :global(p:last-child) {
    margin-bottom: 0;
  }

  .markdown-body :global(ul),
  .markdown-body :global(ol) {
    margin: 0.5rem 0;
    padding-left: 1.5rem;
  }

  .markdown-body :global(li) {
    margin-bottom: 0.25rem;
  }

  .markdown-body :global(pre) {
    background: var(--bg-primary);
    padding: 0.75rem;
    border-radius: 0.5rem;
    overflow-x: auto;
    margin: 0.5rem 0;
    white-space: pre-wrap; /* Wrap long lines */
    word-break: break-all; /* Break long words */
    overflow-wrap: anywhere; /* Ensure break anywhere if needed */
  }

  .markdown-body :global(code) {
    font-family: "Courier New", monospace;
    background: rgba(0, 0, 0, 0.2);
    padding: 0.1rem 0.3rem;
    border-radius: 0.25rem;
    font-size: 0.9em;
  }

  .markdown-body :global(pre) :global(code) {
    background: transparent;
    padding: 0;
    font-size: 0.85rem;
    color: var(--text-secondary);
  }

  .markdown-body :global(a) {
    color: var(--accent);
    text-decoration: none;
    font-weight: 500;
    transition: all 0.2s;
    border-bottom: 1px solid transparent;
  }

  .markdown-body :global(a:hover) {
    border-bottom-color: var(--accent);
    text-shadow: 0 0 8px rgba(59, 130, 246, 0.3);
  }

  /* Make links inside user bubble white */
  .user .markdown-body :global(a) {
    color: white;
    text-decoration: underline;
    opacity: 0.9;
  }

  .user .markdown-body :global(a:hover) {
    opacity: 1;
    text-shadow: none;
  }

  .markdown-body :global(strong) {
    font-weight: 700;
    color: inherit; /* inherit ensures it looks good in user bubbles too */
  }
</style>
