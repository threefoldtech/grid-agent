<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { fade, scale, slide, fly } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import {
    AddProfile,
    UpdateProfile,
    DeleteProfile,
    UpdateAdvancedSettings,
    UpdateGridSettings,
    GetVersion,
    GetAvailableTools,
    SetSafetyMode,
    SetSmartSafetyThreshold,
    ToggleToolOverride,
  } from "../../wailsjs/go/main/App.js";
  import {
    settingsStore,
    activateProfile,
    deactivateProfile,
  } from "../stores/stores";

  const dispatch = createEventDispatcher();
  export let show = false;

  // --- State ---
  let profiles: any[] = [];
  let activeProfileID = "";
  let activeSection = "personas"; // 'personas' | 'config' | 'appearance'

  // Persona Form
  let isCreating = false;
  let editingProfile: any = null;
  let formName = "";
  let formInstructions = "";

  // Advanced Settings
  let advancedProvider = "gemini";
  let advancedModel = "";
  let advancedApiKey = "";
  let isEditingApiKey = false;
  let tempApiKey = "";
  let providerDropdownOpen = false;
  let modelDropdownOpen = false;
  let safetyModeDropdownOpen = false;
  let showApiKey = false;

  // Track original values for change detection
  let originalProvider = "gemini";
  let originalApiKey = "";
  let originalModel = "";

  // Grid Settings
  let gridMnemonics = "";
  let gridNetwork = "";
  let isEditingMnemonic = false;
  let tempMnemonic = "";
  let networkDropdownOpen = false;
  let showMnemonic = false;

  // Track original grid values for change detection
  let originalMnemonics = "";
  let originalNetwork = "";

  // Feedback
  let error = "";
  let successMsg = "";
  let showDeleteModal = false;
  let profileToDelete: any = null;

  // Version
  let appVersion = "";

  // Track modal visibility for latch
  let prevShow = false;

  const availableModels = [
    "gemini-3-pro-preview",
    "gemini-3-flash-preview",
    "gemini-2.5-pro",
    "gemini-2.5-flash",
    "gemini-2.5-flash-lite",
    "gemini-robotics-er-1.5-preview",
  ];

  // --- Reactivity ---
  $: if ($settingsStore) {
    profiles = $settingsStore.profiles || [];
    activeProfileID = $settingsStore.activeProfileID || "";
  }

  // Latch for opening modal
  $: if (show !== prevShow) {
    if (show && settingsStore) {
      resetState();
      if ($settingsStore) {
        advancedApiKey = $settingsStore.geminiApiKey || "";
        advancedModel = $settingsStore.model || "gemini-3-flash-preview";
        enableExportSummary = $settingsStore.enableExportSummary || false;
        // Store original values
        originalApiKey = advancedApiKey;
        originalModel = advancedModel;
        originalEnableExportSummary = enableExportSummary;

        // Load grid settings
        gridMnemonics = $settingsStore.mnemonics || "";
        gridNetwork = $settingsStore.network || "main";
        originalMnemonics = gridMnemonics;
        originalNetwork = gridNetwork;

        // Load safety settings
        safetyMode = $settingsStore.safetyMode || "manual";
        smartSafetyThreshold = $settingsStore.smartSafetyThreshold || "high";
        requireToolApproval = $settingsStore.requireToolApproval || false;
        toolApprovalOverrides = { ...($settingsStore.toolApprovalOverrides || {}) };

        // Load available tools
        if (!toolsLoaded) {
           loadTools();
        }

        // Load version
        GetVersion().then(v => appVersion = v);
      }
    }
    prevShow = show;
  }

  // Detect if config has changed
  let enableExportSummary = false;
  let originalEnableExportSummary = false;

  // Safety settings (moved to separate section)
  let safetyMode = "manual";
  let smartSafetyThreshold = "high";
  let requireToolApproval = false;
  let toolApprovalOverrides: Record<string, boolean> = {};
  let availableTools: Array<{ name: string; description: string }> = [];
  let toolsLoaded = false;
  let isLoadingTools = false;

  $: configHasChanged =
    (advancedProvider !== originalProvider ||
      advancedApiKey !== originalApiKey ||
      advancedModel !== originalModel ||
      enableExportSummary !== originalEnableExportSummary) &&
    (advancedProvider === "ollama" || advancedApiKey.trim() !== "");

  // Detect if grid config has changed
  $: gridConfigHasChanged =
    (gridMnemonics !== originalMnemonics || gridNetwork !== originalNetwork) &&
    gridMnemonics.trim() !== "";

  // --- Actions ---
  function close() {
    dispatch("close");
    resetState();
  }

  function resetState() {
    activeSection = "personas";
    error = "";
    providerDropdownOpen = false;
    modelDropdownOpen = false;
    networkDropdownOpen = false;
    cancelApiKeyEdit();
    cancelEdit();
  }

  function switchSection(section: string) {
    activeSection = section;
    error = "";
    modelDropdownOpen = false;
    cancelEdit();
    if (section === "safety" && !toolsLoaded && !isLoadingTools) {
      loadTools();
    }
  }

  function loadTools() {
    isLoadingTools = true;
    GetAvailableTools().then((tools) => {
      availableTools = tools || [];
      toolsLoaded = true;
    }).catch((err) => {
      console.error("Failed to load available tools:", err);
      availableTools = []; // Ensure array
    }).finally(() => {
      isLoadingTools = false;
    });
  }


    // Close dropdown fetching when clicking outside
  function handleDropdownClickOutside(event: MouseEvent) {
    if (providerDropdownOpen) {
      const target = event.target as HTMLElement;
      if (!target.closest(".custom-select")) {
        providerDropdownOpen = false;
      }
    }
    if (modelDropdownOpen) {
      const target = event.target as HTMLElement;
      if (!target.closest(".custom-select")) {
        modelDropdownOpen = false;
      }
    }
    if (networkDropdownOpen) {
      const target = event.target as HTMLElement;
      if (!target.closest(".custom-select")) {
        networkDropdownOpen = false;
      }
    }
  }

  // --- Persona Logic ---
  function startCreate() {
    isCreating = true;
    editingProfile = null;
    formName = "";
    formInstructions = "";
    error = "";
  }

  function startEdit(profile: any) {
    isCreating = false;
    editingProfile = profile;
    formName = profile.name;
    formInstructions = profile.instructions;
    error = "";
  }

  function cancelEdit() {
    isCreating = false;
    editingProfile = null;
    error = "";
  }

  async function saveProfile() {
    if (!formName.trim()) {
      error = "Name is required";
      return;
    }
    try {
      let newSettings;
      if (isCreating) {
        newSettings = await AddProfile(formName, formInstructions);
      } else {
        newSettings = await UpdateProfile(
          editingProfile.id,
          formName,
          formInstructions,
        );
      }
      settingsStore.set(newSettings);
      cancelEdit();
    } catch (e: any) {
      error = e.message || String(e);
    }
  }

  function confirmDelete(profile: any) {
    profileToDelete = profile;
    showDeleteModal = true;
  }

  async function performDelete() {
    if (!profileToDelete) return;
    try {
      const newSettings = await DeleteProfile(profileToDelete.id);
      settingsStore.set(newSettings);
      showDeleteModal = false;
      profileToDelete = null;
    } catch (e: any) {
      error = e.message || String(e);
    }
  }

  async function toggleActive(id: string) {
    try {
      if (activeProfileID === id) {
        await deactivateProfile();
      } else {
        await activateProfile(id);
      }
    } catch (e: any) {
      error = e.message || String(e);
    }
  }

  // --- Advanced Logic ---
  function startApiKeyEdit() {
    isEditingApiKey = true;
    tempApiKey = "";
  }

  function saveApiKeyEdit() {
    if (tempApiKey.trim()) advancedApiKey = tempApiKey.trim();
    isEditingApiKey = false;
    tempApiKey = "";
  }

  function cancelApiKeyEdit() {
    isEditingApiKey = false;
    tempApiKey = "";
  }

  async function saveAdvanced() {
    if (advancedProvider === "gemini" && !advancedApiKey.trim()) {
      error = "Gemini API Key is required";
      return;
    }
    try {
      const newSettings = await UpdateAdvancedSettings(
        advancedProvider,
        advancedProvider === "gemini" ? advancedApiKey : "http://localhost:11434",
        advancedModel,
        enableExportSummary,
      );
      settingsStore.set(newSettings);
      // Update original values after successful save
      originalProvider = advancedProvider;
      originalApiKey = advancedApiKey;
      originalModel = advancedModel;
      originalEnableExportSummary = enableExportSummary;
      error = "";
    } catch (e: any) {
      error = e.message || String(e);
    }
  }



  // Action Handlers
  async function handleSafetyModeChange() {
    try {
      const newSettings = await SetSafetyMode(safetyMode);
      settingsStore.set(newSettings);
      error = "";
    } catch (e: any) {
      error = e.message || String(e);
    }
  }

  async function handleThresholdChange(newThreshold: string) {
    try {
      smartSafetyThreshold = newThreshold;
      const newSettings = await SetSmartSafetyThreshold(newThreshold);
      settingsStore.set(newSettings);
      error = "";
    } catch (e: any) {
      error = e.message || String(e);
    }
  }

  async function handleToggleOverride(toolName: string) {
    try {
      const newSettings = await ToggleToolOverride(toolName);
      settingsStore.set(newSettings);
      error = "";
    } catch (e: any) {
      error = e.message || String(e);
    }
  }

  // Grid Configuration Functions
  function formatMnemonic(mnemonic: string): string {
    if (!mnemonic) return "";
    const words = mnemonic.trim().split(/\s+/);
    if (words.length < 2) return mnemonic;
    return `${words[0]} ${"*".repeat(20)} ${words[words.length - 1]}`;
  }

  function toggleNetworkDropdown() {
    networkDropdownOpen = !networkDropdownOpen;
  }

  function selectNetwork(network: string) {
    gridNetwork = network;
    networkDropdownOpen = false;
  }

  function startMnemonicEdit() {
    tempMnemonic = ""; // Start with empty box like API key
    isEditingMnemonic = true;
  }

  function saveMnemonicEdit() {
    gridMnemonics = tempMnemonic;
    isEditingMnemonic = false;
  }

  function cancelMnemonicEdit() {
    tempMnemonic = "";
    isEditingMnemonic = false;
  }

  async function saveGridConfig() {
    if (!gridMnemonics.trim()) {
      error = "Mnemonic is required";
      return;
    }
    try {
      const newSettings = await UpdateGridSettings(gridMnemonics, gridNetwork);
      settingsStore.set(newSettings);
      // Update original values after successful save
      originalMnemonics = gridMnemonics;
      originalNetwork = gridNetwork;
      error = "";
    } catch (e: any) {
      error = e.message || String(e);
    }
  }

  function handleModalClick(event: MouseEvent) {
    // Close dropdowns when clicking outside
    if (modelDropdownOpen) {
      modelDropdownOpen = false;
    }
    if (networkDropdownOpen) {
      networkDropdownOpen = false;
    }
  }
</script>

{#if show}
  <div
    class="overlay"
    on:click={close}
    on:keydown={(e) => e.key === "Escape" && close()}
    transition:fade={{ duration: 200 }}
  >
    <div
      class="settings-container"
      on:click|stopPropagation={handleDropdownClickOutside}
      on:keydown={(e) => e.key === "Escape" && close()}
      transition:scale={{ start: 0.96, duration: 300, easing: cubicOut }}
    >
      <!-- Sidebar Navigation -->
      <aside class="settings-sidebar">
        <div class="sidebar-header">
          <h2>Settings</h2>
        </div>
        <nav class="sidebar-nav">
          <button
            class="nav-item"
            class:active={activeSection === "personas"}
            on:click={() => switchSection("personas")}
          >
            <svg
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
              <circle cx="12" cy="7" r="4"></circle>
            </svg>
            <span>Personas</span>
          </button>
          <button
            class="nav-item"
            class:active={activeSection === "config"}
            on:click={() => switchSection("config")}
          >
            <svg
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <circle cx="12" cy="12" r="3"></circle>
              <path
                d="M12 1v6m0 6v6m-9-9h6m6 0h6m-3.636-3.636l-4.243 4.243m0 4.243l4.243 4.243m-8.486 0l4.243-4.243m0-4.243L3.636 3.636"
              ></path>
            </svg>
            <span>AI Configuration</span>
          </button>

          <button
            class="nav-item"
            class:active={activeSection === "grid"}
            on:click={() => switchSection("grid")}
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
            >
              <rect x="3" y="3" width="7" height="7" />
              <rect x="14" y="3" width="7" height="7" />
              <rect x="14" y="14" width="7" height="7" />
              <rect x="3" y="14" width="7" height="7" />
            </svg>
            <span>Grid Configuration</span>
          </button>

          <button
            class="nav-item"
            class:active={activeSection === "safety"}
            on:click={() => switchSection("safety")}
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
            >
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            </svg>
            <span>Safety</span>
          </button>
        </nav>

        <!-- Version Display in Sidebar Footer -->
        {#if appVersion}
          <div class="sidebar-footer">
            <span class="version-badge">{appVersion}</span>
          </div>
        {/if}
      </aside>

      <!-- Main Content Area -->
      <main class="settings-content">
        <button class="close-btn" on:click={close}>
          <svg
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>

        {#if activeSection === "personas"}
          <div class="section-content" in:fade={{ duration: 200 }}>
            {#if isCreating || editingProfile}
              <!-- Persona Editor -->
              <div class="editor-panel" in:slide={{ duration: 200 }}>
                <div class="editor-header">
                  <h3>{isCreating ? "Create Persona" : "Edit Persona"}</h3>
                  <p>Define the identity and behavior of your AI agent.</p>
                </div>

                <div class="form-grid">
                  <div class="input-group">
                    <label for="p-name">Display Name</label>
                    <input
                      id="p-name"
                      type="text"
                      bind:value={formName}
                      placeholder="e.g. Code Reviewer"
                      autocomplete="off"
                    />
                  </div>

                  <div class="input-group full-width">
                    <label for="p-inst">System Instructions</label>
                    <textarea
                      id="p-inst"
                      bind:value={formInstructions}
                      rows="6"
                      placeholder="You are an expert software engineer..."
                    ></textarea>
                  </div>
                </div>

                <div class="editor-actions">
                  <button class="btn secondary" on:click={cancelEdit}
                    >Cancel</button
                  >
                  <button class="btn primary" on:click={saveProfile}>
                    {isCreating ? "Create Persona" : "Save Changes"}
                  </button>
                </div>
              </div>
            {:else}
              <!-- Persona Grid -->
              <div class="section-header">
                <h3>Your Personas</h3>
                <p>Manage your AI agent personalities</p>
              </div>

              {#if profiles.length === 0}
                <div class="empty-state">
                  <svg
                    width="64"
                    height="64"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="1.5"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    class="empty-icon"
                  >
                    <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
                    <circle cx="12" cy="7" r="4"></circle>
                  </svg>
                  <h4>No personas yet</h4>
                  <p>Create your first AI persona to get started</p>
                  <button class="btn primary" on:click={startCreate}>
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
                      <line x1="12" y1="5" x2="12" y2="19"></line>
                      <line x1="5" y1="12" x2="19" y2="12"></line>
                    </svg>
                    Create Persona
                  </button>
                </div>
              {:else}
                <div class="persona-grid">
                  {#each profiles as p (p.id)}
                    <div
                      class="persona-card"
                      class:active={activeProfileID === p.id}
                      transition:scale|local={{ duration: 200 }}
                    >
                      <div class="card-avatar">
                        {p.name[0].toUpperCase()}
                      </div>
                      <h4 class="card-name">{p.name}</h4>
                      {#if activeProfileID !== p.id}
                        <div class="inactive-indicator"></div>
                      {/if}
                      <div class="card-actions">
                        <button
                          class="action-btn"
                          on:click={() => toggleActive(p.id)}
                          title={activeProfileID === p.id
                            ? "Deactivate"
                            : "Activate"}
                        >
                          {#if activeProfileID === p.id}
                            <svg
                              width="18"
                              height="18"
                              viewBox="0 0 24 24"
                              fill="none"
                              stroke="currentColor"
                              stroke-width="2"
                            >
                              <circle cx="12" cy="12" r="10"></circle>
                              <line x1="8" y1="12" x2="16" y2="12"></line>
                            </svg>
                          {:else}
                            <svg
                              width="18"
                              height="18"
                              viewBox="0 0 24 24"
                              fill="none"
                              stroke="currentColor"
                              stroke-width="2"
                            >
                              <circle cx="12" cy="12" r="10"></circle>
                              <polyline points="12 16 16 12 12 8"></polyline>
                              <line x1="8" y1="12" x2="16" y2="12"></line>
                            </svg>
                          {/if}
                        </button>
                        <button
                          class="action-btn"
                          on:click={() => startEdit(p)}
                          title="Edit"
                        >
                          <svg
                            width="18"
                            height="18"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                          >
                            <path d="M12 20h9"></path>
                            <path
                              d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"
                            ></path>
                          </svg>
                        </button>
                        <button
                          class="action-btn delete"
                          on:click={() => confirmDelete(p)}
                          title="Delete"
                        >
                          <svg
                            width="18"
                            height="18"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                          >
                            <polyline points="3 6 5 6 21 6"></polyline>
                            <path
                              d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"
                            ></path>
                          </svg>
                        </button>
                      </div>
                    </div>
                  {/each}
                </div>
              {/if}
            {/if}
          </div>

          <!-- FAB for New Persona -->
          {#if !isCreating && !editingProfile && profiles.length > 0}
            <button
              class="fab"
              on:click={startCreate}
              transition:scale={{ duration: 200 }}
            >
              <svg
                width="24"
                height="24"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2.5"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <line x1="12" y1="5" x2="12" y2="19"></line>
                <line x1="5" y1="12" x2="19" y2="12"></line>
              </svg>
            </button>
          {/if}
        {:else if activeSection === "config"}
          <div class="section-content" in:fade={{ duration: 200 }}>
            <div class="section-header">
              <h3>AI Configuration</h3>
              <p>Configure your AI model and API settings</p>
            </div>

            <div class="config-form">
              <div class="input-group">
                <label for="adv-provider">AI Provider</label>
                <div class="custom-select" class:open={providerDropdownOpen}>
                  <button
                    type="button"
                    class="select-trigger"
                    on:click={() => (providerDropdownOpen = !providerDropdownOpen)}
                  >
                    <span>{advancedProvider === "gemini" ? "Google Gemini" : "Ollama (Local)"}</span>
                    <svg
                      width="12"
                      height="12"
                      viewBox="0 0 12 12"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                      class="select-arrow"
                      class:rotated={providerDropdownOpen}
                    >
                      <path d="M2 4l4 4 4-4" />
                    </svg>
                  </button>
                  {#if providerDropdownOpen}
                    <div
                      class="select-dropdown"
                      transition:slide={{ duration: 200 }}
                    >
                      <button
                        type="button"
                        class="select-option"
                        class:selected={advancedProvider === "gemini"}
                        on:click={() => {
                          advancedProvider = "gemini";
                          // Reset to Gemini models
                          advancedModel = "gemini-3-flash-preview";
                          providerDropdownOpen = false;
                        }}
                      >
                        Google Gemini
                      </button>
                      <button
                        type="button"
                        class="select-option"
                        class:selected={advancedProvider === "ollama"}
                        on:click={() => {
                          advancedProvider = "ollama";
                          // Reset to Ollama models
                          advancedModel = "llama3.1:8b";
                          providerDropdownOpen = false;
                        }}
                      >
                        Ollama (Local)
                      </button>
                    </div>
                  {/if}
                </div>
              </div>

              <div class="input-group">
                <label for="adv-model">AI Model</label>
                <div class="custom-select" class:open={modelDropdownOpen}>
                  <button
                    type="button"
                    class="select-trigger"
                    on:click={() => (modelDropdownOpen = !modelDropdownOpen)}
                  >
                    <span>{advancedModel || "Select a model"}</span>
                    <svg
                      width="12"
                      height="12"
                      viewBox="0 0 12 12"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                      class="select-arrow"
                      class:rotated={modelDropdownOpen}
                    >
                      <path d="M2 4l4 4 4-4" />
                    </svg>
                  </button>
                  {#if modelDropdownOpen}
                    <div
                      class="select-dropdown"
                      transition:slide={{ duration: 200 }}
                    >
                      {#each advancedProvider === "gemini" ? availableModels : [
                        "llama3.1:8b",
                        "llama3.1:70b",
                        "llama3.1:405b",
                        "qwen2.5:7b",
                        "qwen2.5:14b",
                        "qwen2.5:32b",
                        "qwen2.5:72b",
                        "mistral:7b"
                      ] as m}
                        <button
                          type="button"
                          class="select-option"
                          class:selected={advancedModel === m}
                          on:click={() => {
                            advancedModel = m;
                            modelDropdownOpen = false;
                          }}
                        >
                          {m}
                        </button>
                      {/each}
                    </div>
                  {/if}
                </div>
              </div>

              {#if advancedProvider === "gemini"}
                <div class="input-group">
                  <label for="adv-key">Gemini API Key</label>
                  {#if isEditingApiKey}
                    <div class="api-key-edit">
                      <input
                        id="adv-key"
                        type="text"
                        bind:value={tempApiKey}
                        placeholder="Paste API Key here"
                      />
                      <div class="edit-actions">
                        <button
                          class="btn small primary"
                          on:click={saveApiKeyEdit}>OK</button
                        >
                        <button
                          class="btn small secondary"
                          on:click={cancelApiKeyEdit}>Cancel</button
                        >
                      </div>
                    </div>
                  {:else}
                    <div class="api-key-display">
                      <input
                        type="text"
                        value={showApiKey
                          ? (advancedApiKey.length > 8
                            ? advancedApiKey.slice(0, 4) + "****" + advancedApiKey.slice(-4)
                            : "****")
                          : "*".repeat(advancedApiKey.length)}
                        disabled
                      />
                      <button
                        class="btn small outline"
                        on:click={() => showApiKey = !showApiKey}
                      >
                        {showApiKey ? "Hide" : "Show"}
                      </button>
                      <button
                        class="btn small secondary"
                        on:click={startApiKeyEdit}>Change</button
                      >
                    </div>
                  {/if}
                  <p class="helper-text">
                    Your API key is stored securely on your local device.
                  </p>
                </div>
              {/if}

              <div class="input-group">
                <label for="export-summary">Export Options</label>
                <div class="toggle-option">
                  <label class="toggle-label">
                    <input type="checkbox" bind:checked={enableExportSummary} />
                    <span class="toggle-text"
                      >Generate AI Summary on Export</span
                    >
                  </label>
                  <p class="helper-text">
                    When enabled, uses tokens to generate a summary when
                    exporting conversations. Off by default to save tokens.
                  </p>
                </div>
              </div>

              <div class="form-actions">
                <button
                  class="btn primary large"
                  on:click={saveAdvanced}
                  disabled={!configHasChanged}>Save Configuration</button
                >
              </div>
            </div>
          </div>
        {/if}

        <!-- Grid Configuration Section -->
        {#if activeSection === "grid"}
          <div class="section-content">
            <div class="section-header">
              <h3>Grid Configuration</h3>
              <p>Configure your ThreeFold Grid connection</p>
            </div>

            <div class="config-form">
              <!-- Network Dropdown -->
              <div class="input-group">
                <label for="grid-network">Network</label>
                <div class="custom-select" class:open={networkDropdownOpen}>
                  <button
                    type="button"
                    class="select-trigger"
                    on:click={() =>
                      (networkDropdownOpen = !networkDropdownOpen)}
                  >
                    <span>{gridNetwork}</span>
                    <svg
                      class="select-arrow"
                      width="12"
                      height="12"
                      viewBox="0 0 12 12"
                      fill="none"
                    >
                      <path
                        d="M2 4L6 8L10 4"
                        stroke="currentColor"
                        stroke-width="1.5"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                      />
                    </svg>
                  </button>
                  {#if networkDropdownOpen}
                    <div
                      class="select-dropdown"
                      transition:slide={{ duration: 200 }}
                    >
                      {#each ["main", "test", "dev", "qa"] as net}
                        <button
                          type="button"
                          class="select-option"
                          class:selected={gridNetwork === net}
                          on:click={() => {
                            gridNetwork = net;
                            networkDropdownOpen = false;
                          }}
                        >
                          {net}
                        </button>
                      {/each}
                    </div>
                  {/if}
                </div>
              </div>

              <!-- Mnemonic Display/Edit -->
              <div class="input-group">
                <label for="grid-mnemonic">Mnemonic Phrase</label>
                {#if isEditingMnemonic}
                  <div class="api-key-edit">
                    <input
                      id="grid-mnemonic"
                      type="text"
                      bind:value={tempMnemonic}
                      placeholder="Enter your mnemonic phrase"
                    />
                    <div class="edit-actions">
                      <button
                        class="btn small primary"
                        on:click={saveMnemonicEdit}>OK</button
                      >
                      <button
                        class="btn small secondary"
                        on:click={cancelMnemonicEdit}>Cancel</button
                      >
                    </div>
                  </div>
                {:else}
                  <div class="api-key-display">
                    <input
                      type="text"
                      value={showMnemonic
                        ? (() => {
                            const words = gridMnemonics.split(/\s+/);
                            if (words.length > 4) {
                              return words.slice(0, 2).join(" ") + " *** " + words.slice(-2).join(" ");
                            }
                            return "*** ***";
                          })()
                        : "*".repeat(gridMnemonics.split(/\s+/).length)}
                      disabled
                    />
                    <button
                      class="btn small outline"
                      on:click={() => showMnemonic = !showMnemonic}
                    >
                      {showMnemonic ? "Hide" : "Show"}
                    </button>
                    <button
                      class="btn small secondary"
                      on:click={startMnemonicEdit}>Change</button
                    >
                  </div>
                {/if}
                <p class="helper-text">
                  Your mnemonic is stored securely on your local device.
                </p>
              </div>

              <!-- Save Button -->
              <div class="form-actions">
                <button
                  class="btn primary large"
                  on:click={saveGridConfig}
                  disabled={!gridConfigHasChanged}>Save Configuration</button
                >
              </div>
            </div>
          </div>
        {/if}

        <!-- Safety Section -->
        {#if activeSection === "safety"}
          <div class="section-content">
            <div class="section-header">
              <h3>Safety Settings</h3>
              <p>Configure tool execution approval requirements</p>
            </div>

            <div class="config-form">
              <!-- Safety Mode Selector -->
              <div class="setting-item">
                <div class="setting-info">
                  <h4>Safety Mode</h4>
                  <p>Choose how tool execution approvals are handled.</p>
                </div>
                <div class="custom-select" class:open={safetyModeDropdownOpen}>
                  <button 
                    class="select-trigger" 
                    on:click|stopPropagation={() => safetyModeDropdownOpen = !safetyModeDropdownOpen}
                  >
                    <span>
                      {#if safetyMode === 'manual'}
                        Manual Control (Secure)
                      {:else if safetyMode === 'smart'}
                        Smart Guard (AI)
                      {:else if safetyMode === 'turbo'}
                        Turbo Mode (Auto)
                      {/if}
                    </span>
                    <svg
                      class="select-arrow"
                      class:rotated={safetyModeDropdownOpen}
                      width="16"
                      height="16"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                    >
                      <path d="M6 9l6 6 6-6"></path>
                    </svg>
                  </button>
                  
                  {#if safetyModeDropdownOpen}
                    <div class="select-dropdown" transition:slide={{ duration: 150 }}>
                      <button 
                        class="select-option" 
                        class:selected={safetyMode === 'manual'}
                        on:click={() => { safetyMode = 'manual'; handleSafetyModeChange(); safetyModeDropdownOpen = false; }}
                      >
                        Manual Control (Secure)
                      </button>
                      <button 
                        class="select-option" 
                        class:selected={safetyMode === 'smart'}
                        on:click={() => { safetyMode = 'smart'; handleSafetyModeChange(); safetyModeDropdownOpen = false; }}
                      >
                        Smart Guard (AI)
                      </button>
                      <button 
                        class="select-option" 
                        class:selected={safetyMode === 'turbo'}
                        on:click={() => { safetyMode = 'turbo'; handleSafetyModeChange(); safetyModeDropdownOpen = false; }}
                      >
                        Turbo Mode (Auto)
                      </button>
                    </div>
                  {/if}
                </div>
              </div>

            <!-- Conditional UI based on Mode -->
            {#if safetyMode === 'smart'}
              <div class="setting-item">
                <div class="setting-info">
                  <h4>Sensitivity Threshold</h4>
                  <p>Determine how strict the Smart Guard should be.</p>
                </div>
                <div class="segment-control">
                  <button
                    class="segment-btn"
                    class:active={smartSafetyThreshold === 'high'}
                    on:click={() => handleThresholdChange('high')}>
                    Strict
                  </button>
                  <button
                    class="segment-btn"
                    class:active={smartSafetyThreshold === 'medium'}
                    on:click={() => handleThresholdChange('medium')}>
                    Relaxed
                  </button>
                </div>
              </div>
            {/if}

            {#if safetyMode === 'manual'}
              <div class="tools-list">
                {#if isLoadingTools}
                   <div class="loading-tools">Loading tools...</div>
                {:else if availableTools.length === 0}
                  <div class="empty-state">No tools available.</div>
                {:else}
                  {#each availableTools as tool}
                    <div class="tool-item">
                      <div class="tool-info">
                        <span class="tool-name">{tool.name}</span>
                        <span class="tool-desc">{tool.description}</span>
                      </div>
                      <div class="segment-control small">
                        <button
                          class="segment-btn"
                          class:active={toolApprovalOverrides[tool.name] === false}
                          on:click={() => { if (toolApprovalOverrides[tool.name] !== false) handleToggleOverride(tool.name); }}>
                          Auto
                        </button>
                        <button
                          class="segment-btn"
                          class:active={toolApprovalOverrides[tool.name] !== false}
                          on:click={() => { if (toolApprovalOverrides[tool.name] === false) handleToggleOverride(tool.name); }}>
                          Manual
                        </button>
                      </div>

                    </div>
                    {/each}
                  {/if}
                </div>
              {/if}





            </div>
          </div>
        {/if}

        {#if error}
          <div class="error-toast" transition:slide>
            <svg
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="3"
            >
              <circle cx="12" cy="12" r="10"></circle>
              <line x1="12" y1="8" x2="12" y2="12"></line>
              <line x1="12" y1="16" x2="12.01" y2="16"></line>
            </svg>
            {error}
          </div>
        {/if}

      </main>
    </div>
  </div>
{/if}

<!-- Delete Confirmation Modal -->
{#if showDeleteModal}
  <div class="modal-overlay" transition:fade>
    <div class="confirm-dialog" transition:scale={{ start: 0.9 }}>
      <h3>Delete Persona?</h3>
      <p>
        Are you sure you want to delete "<strong>{profileToDelete?.name}</strong
        >"? This action cannot be undone.
      </p>
      <div class="modal-actions">
        <button
          class="btn secondary"
          on:click={() => {
            showDeleteModal = false;
            profileToDelete = null;
          }}>Cancel</button
        >
        <button class="btn danger" on:click={performDelete}>Delete</button>
      </div>
    </div>
  </div>
{/if}

<style>
  /* Overlay */
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(8px);
    z-index: 999;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 2rem;
  }

  /* Main Container */
  .settings-container {
    width: 100%;
    max-width: 1200px;
    height: 85vh;
    max-height: 800px;
    background: var(--bg-primary);
    border-radius: var(--radius-lg);
    display: flex;
    overflow: hidden;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
    border: 1px solid var(--border);
  }

  /* Sidebar */
  .settings-sidebar {
    width: 240px;
    background: var(--bg-secondary);
    border-right: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    flex-shrink: 0;
  }

  .sidebar-header {
    padding: 2rem 1.5rem 1.5rem;
    border-bottom: 1px solid var(--border);
  }

  .sidebar-header h2 {
    margin: 0;
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--text-primary);
  }

  .sidebar-nav {
    padding: 1rem 0.75rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.875rem 1rem;
    background: transparent;
    border: none;
    border-radius: var(--radius-md);
    color: var(--text-secondary);
    font-size: 0.95rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    position: relative;
  }

  .nav-item:hover {
    background: var(--bg-tertiary);
    color: var(--text-primary);
  }

  .nav-item.active {
    background: var(--bg-tertiary);
    color: var(--accent);
  }

  .nav-item.active::before {
    content: "";
    position: absolute;
    left: 0;
    top: 0;
    bottom: 0;
    width: 3px;
    background: var(--accent-gradient);
    border-radius: 0 2px 2px 0;
  }

  .nav-item svg {
    flex-shrink: 0;
  }

  /* Main Content */
  .settings-content {
    flex: 1;
    display: flex;
    flex-direction: column;
    position: relative;
    overflow: hidden;
  }

  .close-btn {
    position: absolute;
    top: 1.5rem;
    right: 1.5rem;
    background: none;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    padding: 0.5rem;
    border-radius: var(--radius-md);
    transition: all 0.2s;
    z-index: 10;
  }

  .close-btn:hover {
    background: var(--bg-tertiary);
    color: var(--text-primary);
  }

  .section-content {
    flex: 1;
    overflow-y: auto;
    padding: 2rem;
  }

  .section-header {
    margin-bottom: 2rem;
  }

  .section-header h3 {
    margin: 0 0 0.5rem 0;
    font-size: 1.75rem;
    font-weight: 700;
    color: var(--text-primary);
  }

  .section-header p {
    margin: 0;
    font-size: 1rem;
    color: var(--text-secondary);
  }

  /* Persona Grid */
  .persona-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: 1.5rem;
  }

  .persona-card {
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: 2rem 1.5rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    transition: all 0.2s;
    position: relative;
  }

  .persona-card:hover {
    border-color: var(--text-secondary);
  }

  .persona-card.active {
    border: 1.5px solid rgba(59, 130, 246, 0.4);
    box-shadow: none;
  }

  .persona-card.active::before {
    content: "";
    position: absolute;
    top: 1rem;
    right: 1rem;
    width: 12px;
    height: 12px;
    background: linear-gradient(135deg, #3b82f6 0%, #06b6d4 100%);
    border-radius: 50%;
    box-shadow:
      0 0 12px rgba(59, 130, 246, 0.8),
      0 0 20px rgba(6, 182, 212, 0.6);
    animation: pulse 2s ease-in-out infinite;
  }

  @keyframes pulse {
    0%,
    100% {
      opacity: 1;
      transform: scale(1);
    }
    50% {
      opacity: 0.7;
      transform: scale(1.1);
    }
  }

  .inactive-indicator {
    position: absolute;
    top: 1rem;
    right: 1rem;
    width: 12px;
    height: 12px;
    background: var(--text-secondary);
    border-radius: 50%;
    opacity: 0.3;
  }

  .card-avatar {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    background: var(--accent-gradient);
    color: white;
    font-size: 1.75rem;
    font-weight: 700;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 1rem;
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
  }

  .card-name {
    margin: 0 0 0.75rem 0;
    font-size: 1.1rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .card-actions {
    display: flex;
    gap: 0.5rem;
    margin-top: auto;
  }

  .action-btn {
    background: var(--bg-tertiary);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    padding: 0.5rem;
    color: var(--text-secondary);
    cursor: pointer;
    transition: all 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .action-btn:hover {
    background: var(--bg-primary);
    color: var(--text-primary);
    border-color: var(--accent);
  }

  .action-btn.delete:hover {
    background: var(--error);
    color: white;
    border-color: var(--error);
  }

  /* FAB */
  .fab {
    position: absolute;
    bottom: 2rem;
    right: 2rem;
    width: 56px;
    height: 56px;
    border-radius: 50%;
    background: var(--accent-gradient);
    border: none;
    color: white;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 8px 16px rgba(59, 130, 246, 0.4);
    transition: all 0.3s;
    z-index: 10;
  }

  .fab:hover {
    transform: scale(1.1);
    box-shadow: 0 12px 24px rgba(59, 130, 246, 0.5);
  }

  /* Empty State */
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 4rem 2rem;
    text-align: center;
  }

  .empty-icon {
    color: var(--text-secondary);
    opacity: 0.4;
    margin-bottom: 1.5rem;
  }

  .empty-state h4 {
    margin: 0 0 0.5rem 0;
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .empty-state p {
    margin: 0 0 2rem 0;
    font-size: 1rem;
    color: var(--text-secondary);
  }

  /* Editor Panel */
  .editor-panel {
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: 2rem;
    position: relative;
    z-index: 20;
  }

  .editor-header {
    margin-bottom: 2rem;
  }

  .editor-header h3 {
    margin: 0 0 0.5rem 0;
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--text-primary);
  }

  .editor-header p {
    margin: 0;
    font-size: 1rem;
    color: var(--text-secondary);
  }

  .form-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: 1.5rem;
    margin-bottom: 2rem;
  }

  .input-group {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .input-group.full-width {
    grid-column: 1 / -1;
  }

  label {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  input,
  textarea {
    background: var(--bg-tertiary);
    border: 1px solid var(--border);
    color: var(--text-primary);
    padding: 0.875rem 1rem;
    border-radius: var(--radius-md);
    font-family: inherit;
    font-size: 1rem;
    width: 100%;
    outline: none;
    transition: all 0.2s;
  }

  input:focus,
  textarea:focus {
    border-color: var(--accent);
    box-shadow: 0 0 0 3px var(--accent-glow, rgba(59, 130, 246, 0.2));
  }

  input:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  textarea {
    resize: vertical;
    min-height: 120px;
  }

  .editor-actions {
    display: flex;
    gap: 1rem;
    justify-content: flex-end;
  }

  /* Config Form */
  .config-form {
    max-width: 600px;
  }

  .config-form .input-group {
    margin-bottom: 2rem;
  }

  /* Custom Select */
  .custom-select {
    position: relative;
    width: 100%;
  }

  .select-trigger {
    width: 100%;
    background: var(--bg-tertiary);
    border: 1px solid var(--border);
    color: var(--text-primary);
    padding: 0.875rem 1rem;
    border-radius: var(--radius-md);
    font-family: inherit;
    font-size: 1rem;
    text-align: left;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: space-between;
    transition: all 0.2s;
  }

  .select-trigger:hover {
    border-color: var(--text-secondary);
  }

  .custom-select.open .select-trigger {
    border-color: var(--accent);
    box-shadow: 0 0 0 3px var(--accent-glow, rgba(59, 130, 246, 0.2));
  }

  .select-arrow {
    transition: transform 0.2s;
    color: var(--text-secondary);
  }

  .select-arrow.rotated {
    transform: rotate(180deg);
  }

  .select-dropdown {
    position: absolute;
    top: calc(100% + 0.5rem);
    left: 0;
    right: 0;
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.3);
    max-height: 300px;
    overflow-y: auto;
    z-index: 1000;
  }

  .select-option {
    width: 100%;
    background: transparent;
    border: none;
    color: var(--text-primary);
    padding: 0.875rem 1rem;
    font-family: inherit;
    font-size: 1rem;
    text-align: left;
    cursor: pointer;
    transition: background 0.15s;
  }

  .select-option:hover {
    background: var(--bg-tertiary);
  }

  .select-option.selected {
    background: var(--accent);
    color: white;
  }

  .select-option.selected:hover {
    background: var(--accent-hover);
  }

  /* API Key */
  .api-key-display,
  .api-key-edit {
    display: flex;
    gap: 0.75rem;
    align-items: flex-start;
  }

  .api-key-display input,
  .api-key-edit input {
    flex: 1;
  }

  .edit-actions {
    display: flex;
    gap: 0.5rem;
  }

  .helper-text {
    font-size: 0.875rem;
    color: var(--text-secondary);
    margin-top: 0.5rem;
    text-align: left;
  }

  /* Segment Control */
  .segment-control {
    display: inline-flex;
    background: var(--bg-tertiary);
    padding: 0.25rem;
    border-radius: var(--radius-md);
    border: 1px solid var(--border);
  }

  .segment-control.small {
    padding: 0.15rem;
  }

  .segment-control.small .segment-btn {
    padding: 0.25rem 0.75rem;
    font-size: 0.8rem;
  }

  .segment-btn {
    padding: 0.5rem 1rem;
    border: none;
    background: transparent;
    color: var(--text-secondary);
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    border-radius: var(--radius-sm);
    transition: all 0.2s;
  }

  .segment-btn:hover {
    color: var(--text-primary);
  }

  .segment-btn.active {
    background: var(--bg-secondary);
    color: var(--accent);
    box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  }

  /* Safety Modes */
  .turbo-warning {
    display: flex;
    gap: 0.75rem;
    padding: 1rem;
    background: rgba(239, 68, 68, 0.1);
    border: 1px solid rgba(239, 68, 68, 0.2);
    border-radius: var(--radius-md);
    color: #ef4444;
    font-size: 0.875rem;
    align-items: flex-start;
    text-align: left;
  }

  .form-actions {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    margin-top: 2rem;
    position: relative;
  }

  /* Buttons */
  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    padding: 0.75rem 1.5rem;
    border-radius: var(--radius-md);
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    border: 1px solid transparent;
    font-size: 0.95rem;
  }

  .btn.small {
    padding: 0.5rem 1rem;
    font-size: 0.875rem;
  }

  .btn.large {
    padding: 1rem 2rem;
    font-size: 1rem;
  }

  .btn.primary {
    background: var(--accent-gradient);
    color: white;
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
  }

  .btn.primary:hover {
    filter: brightness(1.1);
  }

  .btn.primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
    box-shadow: none;
  }

  .btn.primary:disabled:hover {
    transform: none;
    box-shadow: none;
  }

  .btn.secondary {
    background: var(--bg-tertiary);
    color: var(--text-primary);
    border: 1px solid var(--border);
  }

  .btn.secondary:hover {
    background: var(--bg-secondary);
    border-color: var(--accent);
  }

  .btn.outline {
    background: transparent;
    color: var(--text-secondary);
    border: 1px solid var(--border);
  }

  .btn.outline:hover {
    background: var(--bg-tertiary);
    border-color: var(--accent);
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

  /* Error Toast */
  .error-toast {
    position: absolute;
    bottom: 2rem;
    left: 50%;
    transform: translateX(-50%);
    background: var(--error);
    color: white;
    padding: 1rem 1.5rem;
    border-radius: 2rem;
    display: flex;
    align-items: center;
    gap: 0.75rem;
    box-shadow: 0 10px 20px rgba(0, 0, 0, 0.3);
    font-weight: 600;
    font-size: 0.95rem;
    z-index: 100;
  }

  /* Delete Modal */
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 2000;
  }

  .confirm-dialog {
    background: var(--bg-secondary);
    border-radius: var(--radius-lg);
    padding: 2rem;
    max-width: 400px;
    width: 90%;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.3);
    border: 1px solid var(--border);
  }

  .confirm-dialog h3 {
    margin: 0 0 1rem 0;
    color: var(--text-primary);
    font-size: 1.25rem;
    font-weight: 700;
  }

  .confirm-dialog p {
    margin: 0 0 1.5rem 0;
    color: var(--text-secondary);
    line-height: 1.6;
  }

  .modal-actions {
    display: flex;
    gap: 0.75rem;
    justify-content: flex-end;
  }

  /* Scrollbar */
  .section-content::-webkit-scrollbar {
    width: 6px;
  }

  .section-content::-webkit-scrollbar-track {
    background: transparent;
  }

  .section-content::-webkit-scrollbar-thumb {
    background: var(--border);
    border-radius: 3px;
  }

  .section-content::-webkit-scrollbar-thumb:hover {
    background: var(--text-secondary);
  }

  /* Toggle Switch Styles */
  .toggle-option {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .toggle-label {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    cursor: pointer;
    user-select: none;
  }

  .toggle-label input[type="checkbox"] {
    position: relative;
    width: 44px;
    height: 24px;
    appearance: none;
    -webkit-appearance: none;
    background: var(--bg-tertiary);
    border-radius: 24px;
    cursor: pointer;
    transition: all 0.3s ease;
    flex-shrink: 0;
    border: 1px solid var(--border);
  }

  .toggle-label input[type="checkbox"]::before {
    content: "";
    position: absolute;
    top: 50%;
    left: 2px;
    transform: translateY(-50%);
    width: 18px;
    height: 18px;
    background: var(--text-secondary);
    border-radius: 50%;
    transition: all 0.3s ease;
  }

  .toggle-label input[type="checkbox"]:checked {
    background: var(--accent);
    border-color: var(--accent);
  }

  .toggle-label input[type="checkbox"]:checked::before {
    left: 22px;
    background: white;
  }

  .toggle-label input[type="checkbox"]:hover {
    border-color: var(--accent);
  }

  .toggle-text {
    font-size: 0.9rem;
    color: var(--text-primary);
    font-weight: 500;
  }

  /* Simplified Layout Styles */
  .setting-item {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    padding: 1rem 0;
    border-bottom: 1px solid var(--border);
  }

  .setting-info h4 {
    margin: 0 0 0.25rem 0;
    font-size: 1rem;
    font-weight: 500;
    color: var(--text-primary);
  }

  .setting-info p {
    margin: 0;
    font-size: 0.9rem;
    color: var(--text-secondary);
    line-height: 1.5;
    max-width: 90%;
    text-align: left;
  }

  /* Manual Mode Tool List Styles */
  .tool-item {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    padding: 1rem 0;
    border-bottom: 1px solid var(--border);
    gap: 1.5rem;
  }

  .tool-item:last-child {
    border-bottom: none;
  }

  .tool-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    text-align: left;
  }

  .tool-info .tool-name {
    font-weight: 600;
    color: var(--text-primary);
    font-size: 0.95rem;
  }

  .tool-info .tool-desc {
    font-size: 0.85rem;
    color: var(--text-secondary);
    line-height: 1.4;
  }

  /* Flat Tool List Styles */
  .tool-settings-list {
    margin-top: 1.5rem;
  }

  .subsection-title {
    font-size: 0.85rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-secondary);
    margin: 0 0 1rem 0;
    font-weight: 600;
  }

  .flat-tool-list {
    display: flex;
    flex-direction: column;
  }

  .tool-row {
    display: flex;
    align-items: flex-start;
    padding: 0.75rem 0;
    border-bottom: 1px solid var(--border);
    gap: 1rem;
  }

  /* Remove border from last item */
  .tool-row:last-child {
    border-bottom: none;
  }

  .tool-content {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .tool-top {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .tool-name {
    font-weight: 600;
    color: var(--text-primary);
    font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
    font-size: 0.9rem;
  }

  .status-badge {
    font-size: 0.7rem;
    font-weight: 600;
    padding: 0.1rem 0.4rem;
    border-radius: 4px;
    text-transform: uppercase;
  }

  .status-badge.auto {
    background: rgba(16, 185, 129, 0.1);
    color: #10b981;
  }

  .tool-desc {
    font-size: 0.85rem;
    color: var(--text-secondary);
    line-height: 1.4;
    text-align: left;
  }

  /* Toggle Switch Sizing */
  .toggle-switch.small {
    transform: scale(0.9);
    transform-origin: top left;
    margin-top: 2px;
  }

  /* Toggle Switch Adjustments */
  .toggle-switch {
    position: relative;
    display: inline-block;
    width: 44px;
    height: 24px;
    flex-shrink: 0;
  }

  .toggle-switch.large {
    width: 52px;
    height: 28px;
  }

  .toggle-switch.large .toggle-slider {
    border-radius: 28px;
  }
  
  .toggle-switch.large .toggle-slider::before {
    height: 22px;
    width: 22px;
    left: 3px;
    bottom: 3px;
  }

  .toggle-switch.large input:checked + .toggle-slider::before {
    transform: translateX(24px);
  }

  /* Slider Base Styles */
  .toggle-switch input {
    opacity: 0;
    width: 0;
    height: 0;
  }

  .toggle-slider {
    position: absolute;
    cursor: pointer;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: var(--border); /* Default off state */
    transition: 0.3s;
    border-radius: 24px;
  }

  .toggle-slider::before {
    position: absolute;
    content: "";
    height: 18px;
    width: 18px;
    left: 3px;
    bottom: 3px;
    background-color: white;
    transition: 0.3s;
    border-radius: 50%;
    box-shadow: 0 1px 2px rgba(0,0,0,0.2);
  }

  .toggle-switch input:checked + .toggle-slider {
    background-color: var(--accent);
  }

  .toggle-switch input:checked + .toggle-slider::before {
    transform: translateX(20px);
  }


  .toggle-switch input {
    opacity: 0;
    width: 0;
    height: 0;
  }

  .toggle-slider {
    position: absolute;
    cursor: pointer;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: var(--border);
    transition: 0.3s;
    border-radius: 24px;
  }

  .toggle-slider::before {
    position: absolute;
    content: "";
    height: 18px;
    width: 18px;
    left: 3px;
    bottom: 3px;
    background-color: white;
    transition: 0.3s;
    border-radius: 50%;
  }

  .toggle-switch input:checked + .toggle-slider {
    background-color: var(--accent);
  }

  .toggle-switch input:checked + .toggle-slider::before {
    transform: translateX(20px);
  }

  /* Sidebar Footer - Version Display */
  .sidebar-footer {
    margin-top: auto;
    padding: 1rem 1.25rem;
    border-top: 1px solid var(--border);
    text-align: left;
  }

  .version-badge {
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--text-primary);
    background: var(--bg-tertiary);
    padding: 0.35rem 0.75rem;
    border-radius: var(--radius-sm);
    border: 1px solid var(--border);
    display: inline-block;
  }

  /* Light theme override for version badge */
  :global([data-theme="light"]) .version-badge {
    background: var(--bg-tertiary);
    color: var(--text-primary);
    border-color: var(--border);
  }
</style>
