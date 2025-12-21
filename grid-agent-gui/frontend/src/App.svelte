<script lang="ts">
  import { onMount } from "svelte";
  import { fade } from "svelte/transition";
  import {
    GetSettings,
    SaveSettings,
    SendMessage,
    SetTheme,
    AcknowledgeDisclaimer,
  } from "../wailsjs/go/main/App.js";
  import Onboarding from "./components/Onboarding.svelte";
  import ChatInterface from "./components/ChatInterface.svelte";
  import { themeStore, settingsStore, messagesStore } from "./stores/stores";

  import AnsiToHtml from "ansi-to-html";

  let isConfigured = false;
  let isLoading = true;
  let currentTheme = "dark";
  let errorMessage = "";
  let showErrorModal = false;
  let disclaimerAcknowledged = false;
  let isAcknowledging = false;

  // ANSI to HTML converter for error messages
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

  onMount(async () => {
    try {
      const settings = await GetSettings();
      isConfigured = settings.isConfigured;
      disclaimerAcknowledged = settings.disclaimerAcknowledged;
      currentTheme = settings.theme || "dark";
      themeStore.set(currentTheme);
      settingsStore.set(settings);

      // Clear messages on startup (history not persisted across sessions)
      messagesStore.set([]);
    } catch (error) {
      console.error("Failed to load settings:", error);
    } finally {
      isLoading = false;
    }
  });

  async function handleOnboardingComplete(event: CustomEvent) {
    const { mnemonics, network, apiKey } = event.detail;
    try {
      await SaveSettings(mnemonics, network, apiKey);
      isConfigured = true;
      const settings = await GetSettings();
      settingsStore.set(settings);
    } catch (error) {
      console.error("Failed to save settings:", error);
      errorMessage = "Failed to save settings: " + error;
      showErrorModal = true;
    }
  }

  function closeErrorModal() {
    showErrorModal = false;
    errorMessage = "";
  }

  async function handleAcknowledgeDisclaimer() {
    isAcknowledging = true;
    try {
      await AcknowledgeDisclaimer();
      disclaimerAcknowledged = true;
      const settings = await GetSettings();
      settingsStore.set(settings);
    } catch (error) {
      console.error("Failed to acknowledge disclaimer:", error);
      errorMessage = "Failed to save acknowledgment: " + error;
      showErrorModal = true;
    } finally {
      isAcknowledging = false;
    }
  }

  async function toggleTheme() {
    const newTheme = currentTheme === "dark" ? "light" : "dark";
    currentTheme = newTheme;
    themeStore.set(newTheme);
    await SetTheme(newTheme);
  }

  $: document.documentElement.setAttribute("data-theme", currentTheme);
</script>

<main class="app" data-theme={currentTheme}>
  {#if isLoading}
    <div class="loading">
      <div class="spinner"></div>
      <p>Loading Grid Agent...</p>
    </div>
  {:else if !isConfigured}
    <Onboarding on:complete={handleOnboardingComplete} />
  {:else if !disclaimerAcknowledged}
    <!-- Beta Disclaimer Modal -->
    <div class="disclaimer-container">
      <div class="disclaimer-modal" transition:fade>
        <h1 class="disclaimer-title">Early Beta Software</h1>
        <div class="disclaimer-content">
          <p class="disclaimer-intro">
            This application is currently in early beta and may contain bugs or unexpected behavior. Please review the following before proceeding:
          </p>
          <ul class="disclaimer-list">
            <li>This is experimental software under active development</li>
            <li>Responses depend on the AI model and may be unpredictable</li>
            <li>AI models can generate incorrect or misleading information</li>
            <li>Always verify important actions and review commands before execution</li>
          </ul>
        </div>
        <button 
          class="btn disclaimer-btn" 
          on:click={handleAcknowledgeDisclaimer}
          disabled={isAcknowledging}
        >
          {#if isAcknowledging}
            <span class="spinner-small"></span>
            Processing...
          {:else}
            I Understand
          {/if}
        </button>
      </div>
    </div>
  {:else}
    <ChatInterface {toggleTheme} theme={currentTheme} />
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
</main>

<style>
  :global(*) {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
  }

  :global(body) {
    font-family:
      "Inter",
      "Nunito",
      -apple-system,
      BlinkMacSystemFont,
      "Segoe UI",
      Roboto,
      sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    background: var(--bg-primary); /* Ensure background is set on body too */
  }

  :global(:root) {
    /* Richer Dark Theme (Midnight Blue) */
    --bg-primary: #0f172a; /* Deep slate/midnight */
    --bg-secondary: #1e293b; /* Lighter slate */
    --bg-tertiary: #334155; /* Even lighter for borders/elements */
    --bg-glass: rgba(15, 23, 42, 0.85); /* Glass effect base */

    --text-primary: #f8fafc;
    --text-secondary: #94a3b8;

    /* Vibrant Gradient Accent */
    --accent: #3b82f6;
    --accent-gradient: linear-gradient(
      135deg,
      #3b82f6 0%,
      #06b6d4 100%
    ); /* Blue to Cyan */
    --accent-hover: #2563eb;
    --accent-glow: 0 0 15px rgba(59, 130, 246, 0.4);

    --success: #10b981;
    --error: #ef4444;
    --border: #334155;

    --font-heading: "Nunito", sans-serif; /* Use Nunito for headings if loaded */
    --radius-lg: 1rem;
    --radius-md: 0.75rem;
    --radius-sm: 0.5rem;
  }

  :global([data-theme="light"]) {
    --bg-primary: #f8fafc;
    --bg-secondary: #ffffff;
    --bg-tertiary: #e2e8f0;
    --bg-glass: rgba(255, 255, 255, 0.85);

    --text-primary: #0f172a;
    --text-secondary: #64748b;

    /* Slightly softer gradient for light mode */
    --accent: #3b82f6;
    --accent-gradient: linear-gradient(135deg, #2563eb 0%, #0891b2 100%);
    --accent-hover: #1d4ed8;
    --accent-glow: 0 0 10px rgba(37, 99, 235, 0.2);

    --success: #059669;
    --error: #dc2626;
    --border: #cbd5e1;
  }

  .app {
    width: 100vw;
    height: 100vh;
    background: var(--bg-primary);
    color: var(--text-primary);
    overflow: hidden;
  }

  /* Global Utilities */
  :global(.glass) {
    background: var(--bg-glass) !important;
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }

  :global([data-theme="light"] .glass) {
    border-bottom: 1px solid rgba(0, 0, 0, 0.05);
  }

  :global(.text-gradient) {
    background: var(--accent-gradient);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    color: transparent;
  }

  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100vh;
    gap: 1rem;
  }

  .spinner {
    width: 50px;
    height: 50px;
    border: 4px solid var(--border);
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  /* Error Modal */
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
    text-align: center;
  }

  .error-icon {
    font-size: 3rem;
    margin-bottom: 1rem;
  }

  .modal h2 {
    margin: 0 0 1rem 0;
    color: var(--text-primary);
    font-size: 1.25rem;
  }

  .modal p {
    margin: 0 0 1.5rem 0;
    color: var(--text-secondary);
    line-height: 1.5;
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

  .modal-actions {
    display: flex;
    gap: 0.75rem;
    justify-content: center;
  }

  .btn {
    padding: 0.625rem 1.5rem;
    border-radius: 0.5rem;
    font-weight: 600;
    cursor: pointer;
    border: none;
    transition: all 0.2s;
    font-size: 0.875rem;
  }

  .btn.primary {
    background: var(--accent);
    color: white;
  }

  .btn.primary:hover {
    background: var(--accent-hover);
  }

  /* Beta Disclaimer Styles */
  .disclaimer-container {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100vh;
    padding: 2rem;
    background: var(--bg-primary);
  }

  .disclaimer-modal {
    background: var(--bg-secondary);
    border-radius: 1rem;
    padding: 2.5rem 3rem;
    max-width: 540px;
    width: 100%;
    box-shadow: 0 20px 40px -10px rgba(0, 0, 0, 0.4);
    border: 1px solid var(--border);
    text-align: center;
  }

  .disclaimer-title {
    font-size: 1.5rem;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0 0 1.5rem 0;
  }

  .disclaimer-content {
    text-align: left;
    margin-bottom: 2rem;
  }

  .disclaimer-intro {
    font-size: 0.9375rem;
    color: var(--text-secondary);
    line-height: 1.6;
    margin-bottom: 1.5rem;
    text-align: left;
  }

  .disclaimer-list {
    list-style: disc;
    margin: 0;
    padding-left: 1.5rem;
    color: var(--text-secondary);
  }

  .disclaimer-list li {
    font-size: 0.875rem;
    line-height: 1.6;
    margin-bottom: 0.5rem;
  }

  .disclaimer-list li:last-child {
    margin-bottom: 0;
  }

  .disclaimer-btn {
    width: 100%;
    padding: 0.875rem 1.5rem;
    font-size: 0.9375rem;
    font-weight: 500;
    border-radius: 0.5rem;
    border: none;
    background: var(--accent);
    color: white;
    cursor: pointer;
    transition: background 0.15s ease;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
  }

  .disclaimer-btn:hover:not(:disabled) {
    background: var(--accent-hover);
  }

  .disclaimer-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .spinner-small {
    width: 16px;
    height: 16px;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top-color: white;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }
</style>
