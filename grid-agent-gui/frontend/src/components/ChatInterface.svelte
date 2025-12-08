<script lang="ts">
  import { onMount, afterUpdate } from "svelte";
  import {
    SendMessage,
    Logout,
    AbortWorkflow,
    CheckForUpdates,
  } from "../../wailsjs/go/main/App.js";
  import { EventsOn, BrowserOpenURL } from "../../wailsjs/runtime/runtime.js";
  import {
    messagesStore,
    settingsStore,
    deactivateProfile,
  } from "../stores/stores";
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
  let isAborting = false;

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

  // Active Persona Logic
  let activeProfileName = "";
  $: {
    const activeID = $settingsStore.activeProfileID;
    if (activeID && $settingsStore.profiles) {
      const p = $settingsStore.profiles.find((p) => p.id === activeID);
      activeProfileName = p ? p.name : "";
    } else {
      activeProfileName = "";
    }
  }

  async function disableActiveProfile() {
    try {
      await deactivateProfile();
      // Optional: Add a toast notification here
    } catch (err) {
      console.error("Failed to deactivate persona:", err);
      errorMessage = "Failed to disable persona: " + err;
      showErrorModal = true;
    }
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
    if (!currentRequestID || !isSending || isAborting) return;

    isAborting = true;
    try {
      await AbortWorkflow(currentRequestID);
      console.log("Workflow aborted:", currentRequestID);
    } catch (error) {
      console.error("Failed to abort workflow:", error);
      errorMessage = "Failed to abort workflow: " + error;
      showErrorModal = true;
    } finally {
      isAborting = false;
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
    <div class="logo-group">
      <div class="logo">
        <img src={tfLogo} alt="ThreeFold Logo" />
        <span>Grid Agent</span>
      </div>

      {#if activeProfileName}
        <button
          class="persona-badge"
          on:click={disableActiveProfile}
          title="Disable Active Persona"
          transition:fade={{ duration: 200 }}
        >
          <span class="persona-name">{activeProfileName}</span>
          <span class="persona-close">×</span>
        </button>
      {/if}
    </div>

    <div class="controls">
      {#if showUpdateBanner && updateInfo}
        <button
          class="icon-btn update-btn"
          on:click|preventDefault={openUpdateLink}
          disabled={isOpeningLink}
          title="Update Available ({updateInfo.latestVersion})"
        >
          {#if isOpeningLink}
            <svg
              class="animate-spin"
              xmlns="http://www.w3.org/2000/svg"
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              ><path d="M21 12a9 9 0 1 1-6.219-8.56" /></svg
            >
          {:else}
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
          {/if}
        </button>
      {/if}
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
          disabled={isAborting}
          title={isAborting ? "Aborting..." : "Abort workflow"}
        >
          {#if isAborting}
            <svg
              class="animate-spin"
              xmlns="http://www.w3.org/2000/svg"
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              ><path d="M21 12a9 9 0 1 1-6.219-8.56" /></svg
            >
            Aborting...
          {:else}
            <svg
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2.5"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <circle cx="12" cy="12" r="10"></circle>
              <line x1="15" y1="9" x2="9" y2="15"></line>
              <line x1="9" y1="9" x2="15" y2="15"></line>
            </svg>
            Abort
          {/if}
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
    background: radial-gradient(
        circle at 50% 0%,
        rgba(59, 130, 246, 0.08) 0%,
        transparent 50%
      ),
      var(--bg-primary); /* Subtle radial glow */
  }

  header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 1.5rem;
    position: absolute; /* Float header */
    top: 0;
    left: 0;
    right: 0;
    z-index: 10;
    background: var(--bg-glass);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border-bottom: 1px solid var(--border);
    min-height: 4.8rem;
  }

  /* Header Layout */
  .logo-group {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 0.2rem;
  }

  .logo {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .logo span {
    font-weight: 700;
    font-size: 1.1rem;
    letter-spacing: -0.01em;
  }

  /* Active Persona Badge (Subtitle Style) */
  .persona-badge {
    display: flex;
    align-items: center;
    gap: 0.3rem;
    margin-left: 2.5rem; /* Align with text (offset icon) */
    background: transparent;
    border: none;
    padding: 0;
    font-size: 0.75rem;
    color: var(--text-secondary);
    cursor: pointer;
    transition: all 0.2s ease;
    opacity: 0.8;
  }

  .persona-badge:hover {
    background: transparent;
    border: none;
    color: var(--error);
    opacity: 1;
    transform: none;
  }

  .persona-name {
    font-weight: 600;
    max-width: 120px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .persona-close {
    font-size: 1.1em;
    line-height: 0.8;
    margin-left: 0.2rem;
    opacity: 0; /* Hidden by default */
    transform: translateX(-4px);
    transition: all 0.2s ease;
  }

  .persona-badge:hover .persona-close {
    opacity: 1;
    transform: translateX(0);
  }

  /* Light Theme Overrides for Persona Badge */
  :global([data-theme="light"]) .persona-badge {
    color: #64748b; /* Slate-500 */
  }

  :global([data-theme="light"]) .persona-badge:hover {
    color: #ef4444; /* Red-500 */
  }
  .logo img {
    height: 28px;
    width: auto;
    border-radius: 6px;
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

  /* Floating Input Area */
  .input-area {
    padding: 1.5rem;
    background: transparent; /* Transparent to let gradient show */
    z-index: 10;
  }

  .input-wrapper {
    display: flex;
    gap: 0.75rem;
    background: var(--bg-secondary);
    padding: 0.5rem;
    border-radius: var(--radius-lg);
    border: 1px solid var(--border);
    box-shadow:
      0 4px 6px -1px rgba(0, 0, 0, 0.1),
      0 2px 4px -1px rgba(0, 0, 0, 0.06);
    transition: all 0.2s ease;
    align-items: flex-end; /* Align bottom for multiline */
  }

  .input-wrapper:focus-within {
    border-color: var(--accent);
    box-shadow: var(--accent-glow);
  }

  textarea {
    flex: 1;
    background: transparent;
    border: none;
    padding: 0.75rem 1rem;
    color: var(--text-primary);
    font-size: 1rem;
    resize: none;
    max-height: 150px;
    font-family: inherit;
  }

  textarea:focus {
    outline: none;
  }

  /* Gradient Send Button */
  .send-btn,
  .abort-btn-input {
    padding: 0.75rem 1.5rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
    white-space: nowrap;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .send-btn {
    border-radius: var(--radius-md);
    border: none;
    color: white;
    background: var(--accent-gradient);
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
  }

  .send-btn:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 6px 16px rgba(59, 130, 246, 0.4);
    filter: brightness(1.1);
  }

  .send-btn:active:not(:disabled) {
    transform: translateY(0);
  }

  .send-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    background: var(--bg-tertiary);
    color: var(--text-secondary);
    box-shadow: none;
  }

  .abort-btn-input {
    background: rgba(127, 29, 29, 0.3);
    color: #fca5a5;
    border: 1px solid rgba(153, 27, 27, 0.5);
    border-radius: 0.75rem;
    transition: all 0.2s;
  }

  .abort-btn-input:hover:not(:disabled) {
    background: rgba(153, 27, 27, 0.4);
    color: #fecaca;
    border-color: #991b1b;
  }

  /* Messages Container Adjustment for Fixed Header */
  .messages {
    flex: 1;
    overflow-y: auto;
    padding: 1rem 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    padding-top: 6rem; /* Account for fixed header */
    scroll-behavior: smooth;
  }

  /* Ensure scrollbar looks good in dark mode */
  .messages::-webkit-scrollbar {
    width: 8px;
  }

  .messages::-webkit-scrollbar-track {
    background: transparent;
  }

  .messages::-webkit-scrollbar-thumb {
    background-color: var(--bg-tertiary);
    border-radius: 4px;
  }

  /* Empty State */
  .empty-state {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    color: var(--text-secondary);
    opacity: 0.8;
  }

  .empty-state h2 {
    font-size: 2rem;
    margin-bottom: 0.5rem;
    background: var(--accent-gradient);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    font-weight: 800;
    letter-spacing: -0.02em;
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
    background-color: #7f1d1d;
    color: #fca5a5;
    border: 1px solid #991b1b;
  }

  .btn.danger:hover {
    background-color: #991b1b;
    color: #fecaca;
  }

  /* Error Modal */
  .error-modal {
    text-align: center;
  }

  .error-icon {
    font-size: 3rem;
    margin-bottom: 1rem;
  }

  .logo span {
    font-size: 1.35rem;
    font-weight: 800;
    background: var(--accent-gradient);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    letter-spacing: -0.02em;
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

  /* Update Button Flashing Animation */
  @keyframes flash-accent {
    0%,
    100% {
      background: transparent;
      color: var(--text-secondary);
    }
    50% {
      background: var(--accent);
      color: white;
    }
  }

  .update-btn {
    animation: flash-accent 2s infinite;
  }
</style>
