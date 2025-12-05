<script lang="ts">
  import { onMount, afterUpdate } from "svelte";
  import {
    SendMessage,
    Logout,
    AbortWorkflow,
    CheckForUpdates,
  } from "../../wailsjs/go/main/App.js";
  import { EventsOn, BrowserOpenURL } from "../../wailsjs/runtime/runtime.js";
  import { messagesStore, settingsStore } from "../stores/stores";
  import ChatMessage from "./ChatMessage.svelte";
  import Settings from "./Settings.svelte";
  import { GenerateSummary, ExportChat } from "../../wailsjs/go/main/App.js";
  import { fade, fly } from "svelte/transition";
  import tfLogo from "../assets/images/tf-logo.png";
  import AnsiToHtml from "ansi-to-html";

  export let toggleTheme: () => void;
  export let theme: string;

  let input = "";
  let chatContainer: HTMLElement;
  let isSending = false;
  let currentRequestID = "";
  let showLogoutModal = false;
  let showErrorModal = false;
  let showSettings = false;
  let errorMessage = "";
  let isExporting = false;

  // Update notification state
  let showUpdateBanner = false;
  let updateInfo: { latestVersion: string; releaseURL: string } | null = null;
  let isOpeningLink = false;

  function openUpdateLink() {
    if (isOpeningLink || !updateInfo) return;
    isOpeningLink = true;
    BrowserOpenURL(updateInfo.releaseURL);
    setTimeout(() => (isOpeningLink = false), 2000);
  }

  // ANSI to HTML converter
  const ansiConverter = new AnsiToHtml({
    fg: "#d4d4d4",
    bg: "#1e1e1e",
    newline: true,
    escapeXML: true,
  });

  function renderAnsi(text: string): string {
    if (!text) return "";
    return ansiConverter.toHtml(text);
  }

  function scrollToBottom() {
    if (chatContainer) {
      chatContainer.scrollTop = chatContainer.scrollHeight;
    }
  }

  afterUpdate(scrollToBottom);

  async function handleSubmit() {
    if (!input.trim() || isSending) return;

    const requestID = `req_${Date.now()}_${Math.random()}`;

    const placeholderMsg = {
      role: "agent",
      content: "",
      timestamp: new Date().toISOString(),
      requestID: requestID,
      steps: [
        {
          progressText: "🤔 Processing your request...",
          exportPrefix: "",
          content: "Processing your request...",
          output: "",
          error: "",
        },
      ],
    };

    const userMsg = {
      role: "user",
      content: input,
      timestamp: new Date().toISOString(),
      isCommand: false,
      output: "",
      error: "",
    };

    messagesStore.update((msgs) => [...msgs, userMsg, placeholderMsg]);

    const messageToSend = input;
    input = "";
    isSending = true;
    currentRequestID = requestID; // Track current request

    try {
      const response = await SendMessage(messageToSend, requestID);
      messagesStore.update((msgs) => {
        const index = msgs.findIndex((m) => m.requestID === response.requestID);
        if (index !== -1) {
          const newMsgs = [...msgs];
          newMsgs[index] = response;
          return newMsgs;
        }
        return [...msgs, response]; // Fallback
      });
    } catch (error) {
      console.error("Failed to send message:", error);
      // Preserve existing steps and add error
      messagesStore.update((msgs) => {
        const index = msgs.findIndex((m) => m.requestID === requestID);
        if (index !== -1) {
          const existingMsg = msgs[index];
          // Keep existing steps (minus placeholder), add error step
          const steps = (existingMsg.steps || []).filter(
            (s) => !s.progressText?.includes("🤔 Processing"),
          );
          steps.push({
            progressText: "❌ Error",
            exportPrefix: "",
            content: "",
            output: "",
            error: String(error),
          });
          msgs[index] = { ...existingMsg, content: "Error: " + error, steps };
        }
        return [...msgs];
      });
    } finally {
      isSending = false;
      currentRequestID = ""; // Clear current request
    }
  }

  async function handleAbort() {
    if (!currentRequestID || !isSending) return;

    try {
      await AbortWorkflow(currentRequestID);
      console.log("Workflow aborted:", currentRequestID);
    } catch (error) {
      console.error("Failed to abort workflow:", error);
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSubmit();
    }
  }

  function handleLogout() {
    showLogoutModal = true;
  }

  function cancelLogout() {
    showLogoutModal = false;
  }

  async function confirmLogout() {
    showLogoutModal = false;
    try {
      await Logout();
      // Clear messages
      messagesStore.set([]);
      // Reload to show onboarding
      window.location.reload();
    } catch (error) {
      console.error("Logout failed:", error);
      errorMessage = "Failed to logout: " + error;
      showErrorModal = true;
    }
  }

  function closeErrorModal() {
    showErrorModal = false;
    errorMessage = "";
  }

  function toggleSettings() {
    showSettings = !showSettings;
  }

  async function handleExport() {
    if (isExporting || $messagesStore.length === 0) return;
    isExporting = true;

    try {
      // Format chat history for summary generation
      const historyText = $messagesStore
        .map((msg) => `${msg.role.toUpperCase()}: ${msg.content}`)
        .join("\n\n");

      // Generate summary
      const summary = await GenerateSummary(historyText);

      // Format full export content
      const date = new Date().toLocaleString();
      const network = $settingsStore?.network || "Unknown";

      let exportContent = `# Grid Agent Conversation\n\n`;
      exportContent += `**Date:** ${date}\n`;
      exportContent += `**Network:** ${network}\n\n`;
      exportContent += `## Summary\n${summary}\n\n`;
      exportContent += `## Full Conversation\n\n`;

      $messagesStore.forEach((msg) => {
        const roleIcon = msg.role === "user" ? "👤" : "🤖";
        const time = new Date(msg.timestamp).toLocaleTimeString();
        exportContent += `### ${roleIcon} ${msg.role.toUpperCase()} (${time})\n\n`;
        exportContent += `${msg.content}\n\n`;

        if (msg.steps && msg.steps.length > 0) {
          exportContent += `#### 🛠️ Workflow Steps\n\n`;
          msg.steps.forEach((step, index) => {
            exportContent += `**${index + 1}. ${step.progressText}**\n\n`;

            // Content based on exportPrefix
            if (step.exportPrefix) {
              exportContent += `${step.exportPrefix} ${step.content}\n`;
            } else {
              exportContent += `${step.content}\n`;
            }

            // Output in code block
            if (step.output) {
              exportContent += `\n\`\`\`bash\n${step.output.trim()}\n\`\`\`\n`;
            }

            // Error in alert block
            if (step.error) {
              exportContent += `\n> [!CAUTION]\n> **Error**: ${step.error}\n`;
            }
            exportContent += `\n`;
          });
        }

        exportContent += `---\n\n`;
      });

      // Save file
      await ExportChat(exportContent, `grid-agent-chat-${Date.now()}.md`);
    } catch (error) {
      console.error("Export failed:", error);
      errorMessage = "Failed to export chat: " + error;
      showErrorModal = true;
    } finally {
      isExporting = false;
    }
  }

  // Set up real-time event listeners
  onMount(() => {
    console.log("[DEBUG] Setting up event listeners");

    // Check for updates on startup
    CheckForUpdates()
      .then((info) => {
        if (info.updateAvailable) {
          updateInfo = {
            latestVersion: info.latestVersion,
            releaseURL: info.releaseURL,
          };
          showUpdateBanner = true;
        }
      })
      .catch((err) => {
        console.log("Failed to check for updates:", err);
      });

    // Listen for real-time command output
    EventsOn(
      "command-output",
      (data: {
        requestID: string;
        commandID: string;
        line: string;
        type: string;
      }) => {
        console.log("[DEBUG] Received command-output event:", data);

        // Update the store by accessing current value
        messagesStore.update((messages) => {
          console.log("[DEBUG] Updating messages store with command output");

          // Find the specific agent message by requestID
          const targetAgentMessage = messages.find(
            (m) => m.requestID === data.requestID,
          );

          if (!targetAgentMessage) {
            console.log(
              `[DEBUG] No agent message found for requestID ${data.requestID} - skipping real-time update`,
            );
            return messages; // Don't update if no matching message found
          }

          // Ensure steps array exists
          if (!targetAgentMessage.steps) {
            targetAgentMessage.steps = [];
          }

          // Find the command step by commandID (precise matching)
          let commandStep = targetAgentMessage.steps.find(
            (s) => s.commandID === data.commandID,
          );

          if (!commandStep) {
            console.log(
              `[DEBUG] No command step found for commandID ${data.commandID} - skipping`,
            );
            return messages;
          }

          // Clear any placeholder like 'thinking'
          targetAgentMessage.steps = targetAgentMessage.steps.filter(
            (s) => !s.progressText?.includes("🤔 Processing"),
          );

          const lastCommandStep = commandStep;
          if (!lastCommandStep) {
            console.log(
              "[DEBUG] No agent message with command steps found - skipping real-time update",
            );
            return messages; // Don't update if no matching message found
          }

          // Update the existing command step's output in real-time
          if (!lastCommandStep.output) {
            lastCommandStep.output = "";
          }
          lastCommandStep.output += data.line + "\n";
          console.log(
            "[DEBUG] Updated command step output:",
            lastCommandStep.output,
          );

          // Return updated messages array
          return [...messages];
        });
      },
    );

    // Listen for agent progress events
    EventsOn("agent-progress", (event: { requestID: string; step: any }) => {
      const { requestID, step } = event;
      console.log("[DEBUG] Received agent-progress event:", event);

      messagesStore.update((messages) => {
        const targetAgentMessage = messages.find(
          (m) => m.requestID === requestID,
        );

        if (targetAgentMessage) {
          // Clear any placeholder like 'thinking'
          targetAgentMessage.steps = targetAgentMessage.steps.filter(
            (s) => !s.progressText?.includes("🤔 Processing"),
          );
          if (!targetAgentMessage.steps) {
            targetAgentMessage.steps = [];
          }

          // Update existing step or add new one
          const existingStepIndex = targetAgentMessage.steps.findIndex(
            (s) =>
              s.progressText === step.progressText &&
              s.content === step.content,
          );

          if (existingStepIndex >= 0) {
            // Update existing step
            targetAgentMessage.steps[existingStepIndex] = step;
          } else {
            // Add new step
            targetAgentMessage.steps.push(step);
          }
        }

        return [...messages];
      });
    });

    console.log("[DEBUG] Event listeners set up complete");
  });
</script>

<div class="chat-interface" in:fade>
  <header>
    <div class="logo">
      <img src={tfLogo} alt="ThreeFold Logo" />
      <span>Grid Agent</span>
    </div>

    <!-- Update Banner -->
    {#if showUpdateBanner && updateInfo}
      <div class="update-banner" transition:fly={{ y: -20, duration: 400 }}>
        <div class="banner-content">
          <div class="icon-wrapper">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path
                d="M4.5 16.5c-1.5 1.26-2 5-2 5s3.74-.5 5-2c.71-.84.7-2.13-.09-2.91a2.18 2.18 0 0 0-2.91-.09z"
              ></path>
              <path
                d="m12 15-3-3a22 22 0 0 1 2-3.95A12.88 12.88 0 0 1 22 2c0 2.72-.78 7.5-6 11a22.35 22.35 0 0 1-4 2z"
              ></path>
              <path d="M9 12H4s.55-3.03 2-4c1.62-1.08 5 0 5 0"></path>
              <path d="M12 15v5s3.03-.55 4-2c1.08-1.62 0-5 0-5"></path>
            </svg>
          </div>
          <div class="text-group">
            <span class="banner-title">Update Available</span>
            <span class="banner-desc"
              >Version <strong>{updateInfo.latestVersion}</strong> is closer than
              you think</span
            >
          </div>
        </div>
        <div class="banner-actions">
          <button
            class="btn-update"
            on:click|preventDefault={openUpdateLink}
            disabled={isOpeningLink}
          >
            {#if isOpeningLink}
              Opening...
            {:else}
              Download Update
            {/if}
          </button>
          <button
            class="btn-dismiss"
            on:click={() => (showUpdateBanner = false)}
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              ><line x1="18" y1="6" x2="6" y2="18"></line><line
                x1="6"
                y1="6"
                x2="18"
                y2="18"
              ></line></svg
            >
          </button>
        </div>
      </div>
    {/if}

    <div class="controls">
      <button
        class="icon-btn"
        on:click={handleExport}
        title="Export Conversation"
        disabled={isExporting}
      >
        {#if isExporting}
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="animate-spin"><path d="M21 12a9 9 0 1 1-6.219-8.56" /></svg
          >
        {:else}
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            ><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" /><polyline
              points="7 10 12 15 17 10"
            /><line x1="12" x2="12" y1="15" y2="3" /></svg
          >
        {/if}
      </button>
      <button class="icon-btn" on:click={toggleSettings} title="Settings">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="20"
          height="20"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          ><path
            d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.09a2 2 0 0 1-1-1.74v-.51a2 2 0 0 1 1-1.72l.15-.1a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"
          /><circle cx="12" cy="12" r="3" /></svg
        >
      </button>
      <button class="icon-btn" on:click={toggleTheme} title="Toggle Theme">
        {#if theme === "dark"}
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            ><path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z" /></svg
          >
        {:else}
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            ><circle cx="12" cy="12" r="4" /><path d="M12 2v2" /><path
              d="M12 20v2"
            /><path d="m4.93 4.93 1.41 1.41" /><path
              d="m17.66 17.66 1.41 1.41"
            /><path d="M2 12h2" /><path d="M20 12h2" /><path
              d="m6.34 17.66-1.41 1.41"
            /><path d="m19.07 4.93-1.41 1.41" /></svg
          >
        {/if}
      </button>
      <button
        class="icon-btn logout-btn"
        on:click={handleLogout}
        title="Logout"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="20"
          height="20"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          ><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" /><polyline
            points="16 17 21 12 16 7"
          /><line x1="21" x2="9" y1="12" y2="12" /></svg
        >
      </button>
    </div>
  </header>

  <div class="messages" bind:this={chatContainer}>
    {#if $messagesStore.length === 0}
      <div class="empty-state">
        <h2>How can I help you today?</h2>
        <p>Ask me to deploy VMs, check nodes, or manage your grid resources.</p>
      </div>
    {/if}

    {#each $messagesStore as msg}
      <ChatMessage message={msg} />
    {/each}
  </div>

  <div class="input-area">
    <div class="input-wrapper">
      <textarea
        bind:value={input}
        on:keydown={handleKeydown}
        placeholder="Type a message..."
        rows="1"
      ></textarea>
      {#if isSending}
        <button
          class="abort-btn-input"
          on:click={handleAbort}
          title="Abort workflow"
        >
          ⏹ Abort
        </button>
      {:else}
        <button
          class="send-btn"
          on:click={handleSubmit}
          disabled={!input.trim() || isSending}
        >
          Send
        </button>
      {/if}
    </div>
  </div>
</div>

<!-- Logout Confirmation Modal -->
{#if showLogoutModal}
  <div class="modal-overlay" on:click={cancelLogout} transition:fade>
    <div class="modal" on:click|stopPropagation transition:fade>
      <h2>Confirm Logout</h2>
      <p>Are you sure you want to logout?</p>
      <p class="warning">
        This will clear your credentials and return to the setup screen.
      </p>
      <div class="modal-actions">
        <button class="btn secondary" on:click={cancelLogout}>Cancel</button>
        <button class="btn danger" on:click={confirmLogout}>Logout</button>
      </div>
    </div>
  </div>
{/if}

<!-- Error Modal -->
{#if showErrorModal}
  <div class="modal-overlay" on:click={closeErrorModal} transition:fade>
    <div class="modal error-modal" on:click|stopPropagation transition:fade>
      <div class="error-icon">⚠️</div>
      <h2>Error</h2>
      <p class="error-text">{@html renderAnsi(errorMessage)}</p>
      <div class="modal-actions">
        <button class="btn primary" on:click={closeErrorModal}>OK</button>
      </div>
    </div>
  </div>
{/if}

<Settings show={showSettings} on:close={() => (showSettings = false)} />

<style>
  .chat-interface {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--bg-primary);
  }

  header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 1.5rem;
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border);
  }

  .logo {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    font-weight: 600;
    font-size: 1.125rem;
  }

  .logo img {
    height: 24px;
    width: auto;
  }

  .controls {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .icon-btn {
    background: transparent;
    border: none;
    cursor: pointer;
    font-size: 1.25rem;
    padding: 0.5rem;
    border-radius: 0.5rem;
    transition: all 0.2s ease;
    width: 2.5rem;
    height: 2.5rem;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-secondary);
    position: relative;
  }

  .icon-btn:hover {
    background: var(--bg-tertiary);
    color: var(--text-primary);
    transform: translateY(-1px);
  }

  .icon-btn:active {
    transform: translateY(0);
  }

  .icon-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .icon-btn:disabled:hover {
    background: transparent;
    transform: none;
  }

  .logout-btn {
    color: var(--text-secondary);
  }

  .logout-btn:hover {
    color: var(--error);
    background: rgba(239, 68, 68, 0.1);
  }

  .messages {
    flex: 1;
    overflow-y: auto;
    padding: 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .empty-state {
    text-align: center;
    margin-top: 20vh;
    color: var(--text-secondary);
  }

  .empty-state h2 {
    color: var(--text-primary);
    margin-bottom: 0.5rem;
  }

  .input-area {
    padding: 1.5rem;
    background: var(--bg-secondary);
    border-top: 1px solid var(--border);
  }

  .input-wrapper {
    display: flex;
    gap: 1rem;
    background: var(--bg-primary);
    padding: 0.75rem;
    border-radius: 0.75rem;
    border: 1px solid var(--border);
    transition: border-color 0.2s;
  }

  .input-wrapper:focus-within {
    border-color: var(--accent);
  }

  textarea {
    flex: 1;
    background: transparent;
    border: none;
    color: var(--text-primary);
    font-size: 1rem;
    resize: none;
    padding: 0.25rem;
    font-family: inherit;
  }

  textarea:focus {
    outline: none;
  }

  .send-btn {
    background: var(--accent);
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 0.5rem;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.2s;
  }

  .send-btn:hover:not(:disabled) {
    background: var(--accent-hover);
  }

  .send-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .abort-btn-input {
    background: var(--error);
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 0.5rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    display: flex;
    align-items: center;
    gap: 0.25rem;
  }

  .abort-btn-input:hover {
    background: #dc2626;
    transform: translateY(-1px);
  }

  .update-banner {
    background: rgba(29, 78, 216, 0.15);
    backdrop-filter: blur(8px);
    border: 1px solid rgba(59, 130, 246, 0.3);
    margin: 1rem 1.5rem 0;
    padding: 0.75rem 1rem;
    border-radius: 12px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
    box-shadow:
      0 4px 6px -1px rgba(0, 0, 0, 0.1),
      0 2px 4px -1px rgba(0, 0, 0, 0.06);
  }

  .banner-content {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .icon-wrapper {
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    width: 36px;
    height: 36px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    box-shadow: 0 2px 4px rgba(37, 99, 235, 0.3);
  }

  .text-group {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 0.1rem;
  }

  .banner-title {
    font-size: 0.85rem;
    font-weight: 700;
    color: #93c5fd;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .banner-desc {
    font-size: 0.95rem;
    color: #e2e8f0;
  }

  .banner-desc strong {
    color: white;
    font-weight: 600;
  }

  .banner-actions {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .btn-update {
    background: linear-gradient(90deg, #2563eb 0%, #1d4ed8 100%);
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 8px;
    font-size: 0.9rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    box-shadow: 0 2px 4px rgba(37, 99, 235, 0.2);
    display: flex;
    align-items: center;
    justify-content: center;
    min-width: 140px;
  }

  .btn-update:hover:not(:disabled) {
    box-shadow: 0 4px 6px rgba(37, 99, 235, 0.3);
    background: linear-gradient(90deg, #3b82f6 0%, #2563eb 100%);
  }

  .btn-update:active:not(:disabled) {
    transform: translateY(1px);
    box-shadow: 0 1px 2px rgba(37, 99, 235, 0.2);
  }

  .btn-update:disabled {
    opacity: 0.7;
    cursor: wait;
    transform: none;
  }

  .btn-dismiss {
    background: transparent;
    border: none;
    color: #94a3b8;
    padding: 0.4rem;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .btn-dismiss:hover {
    background: rgba(255, 255, 255, 0.1);
    color: white;
  }

  /* Logout Modal */
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal {
    background: var(--bg-secondary);
    border-radius: 1rem;
    padding: 2rem;
    max-width: 400px;
    width: 90%;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.3);
    border: 1px solid var(--border);
  }

  .modal h2 {
    margin: 0 0 1rem 0;
    color: var(--text-primary);
    font-size: 1.25rem;
  }

  .modal p {
    margin: 0 0 0.5rem 0;
    color: var(--text-secondary);
    line-height: 1.5;
  }

  .modal .warning {
    color: var(--error);
    font-size: 0.875rem;
    margin-bottom: 1.5rem;
  }

  .modal-actions {
    display: flex;
    gap: 0.75rem;
    justify-content: flex-end;
  }

  .btn {
    padding: 0.625rem 1.25rem;
    border-radius: 0.5rem;
    font-weight: 600;
    cursor: pointer;
    border: none;
    transition: all 0.2s;
    font-size: 0.875rem;
  }

  .btn.secondary {
    background: transparent;
    color: var(--text-secondary);
    border: 1px solid var(--border);
  }

  .btn.secondary:hover {
    background: var(--bg-tertiary);
    color: var(--text-primary);
  }

  .btn.danger {
    background: var(--error);
    color: white;
  }

  .btn.danger:hover {
    background: #dc2626;
  }

  /* Error Modal */
  .error-modal {
    text-align: center;
  }

  .error-icon {
    font-size: 3rem;
    margin-bottom: 1rem;
  }

  .error-text {
    font-family: "Courier New", Consolas, Monaco, monospace;
    font-size: 0.875rem;
    text-align: left;
    background: var(--bg-primary);
    padding: 1rem;
    border-radius: 0.5rem;
    overflow-x: auto;
  }

  .btn.primary {
    background: var(--accent);
    color: white;
  }

  .btn.primary:hover {
    background: var(--accent-hover);
  }

  .animate-spin {
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }

  /* Light Theme Overrides for Update Banner */
  :global([data-theme="light"]) .update-banner {
    background: rgba(59, 130, 246, 0.1);
    border: 1px solid rgba(59, 130, 246, 0.2);
  }

  :global([data-theme="light"]) .banner-title {
    color: #1e40af; /* blue-800 */
  }

  :global([data-theme="light"]) .banner-desc {
    color: #475569; /* slate-600 */
  }

  :global([data-theme="light"]) .banner-desc strong {
    color: #1e3a8a; /* blue-900 */
  }

  :global([data-theme="light"]) .btn-dismiss {
    color: #64748b; /* slate-500 */
  }

  :global([data-theme="light"]) .btn-dismiss:hover {
    color: #0f172a; /* slate-900 */
    background: rgba(0, 0, 0, 0.05); /* slightly dark hover */
  }
</style>
