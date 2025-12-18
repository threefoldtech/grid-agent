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
  let advancedModel = "";
  let advancedApiKey = "";
  let isEditingApiKey = false;
  let tempApiKey = "";
  let modelDropdownOpen = false;
  let showApiKey = false;

  // Track original values for change detection
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

        // Load version
        GetVersion().then(v => appVersion = v);
      }
    }
    prevShow = show;
  }

  // Detect if config has changed
  let enableExportSummary = false;
  let originalEnableExportSummary = false;

  $: configHasChanged =
    (advancedApiKey !== originalApiKey ||
      advancedModel !== originalModel ||
      enableExportSummary !== originalEnableExportSummary) &&
    advancedApiKey.trim() !== "";

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
  }

  // Close dropdown when clicking outside
  function handleDropdownClickOutside(event: MouseEvent) {
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
    if (!advancedApiKey.trim()) {
      error = "API Key is required";
      return;
    }
    try {
      const newSettings = await UpdateAdvancedSettings(
        advancedApiKey,
        advancedModel,
        enableExportSummary,
      );
      settingsStore.set(newSettings);
      // Update original values after successful save
      originalApiKey = advancedApiKey;
      originalModel = advancedModel;
      originalEnableExportSummary = enableExportSummary;
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
            class:active={activeSection === "docs"}
            on:click={() => switchSection("docs")}
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
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
              <polyline points="14,2 14,8 20,8"></polyline>
              <line x1="16" y1="13" x2="8" y2="13"></line>
              <line x1="16" y1="17" x2="8" y2="17"></line>
              <polyline points="10,9 9,9 8,9"></polyline>
            </svg>
            <span>Docs</span>
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
                      {#each availableModels as m}
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
        {:else if activeSection === "docs"}
          <div class="section-content" in:fade={{ duration: 200 }}>
            <div class="section-header">
              <h3>Documentation</h3>
              <p>Learn how to use Grid Agent effectively</p>
            </div>

            <div class="docs-content">
              <!-- Getting Started -->
              <div class="docs-section">
                <h4>🚀 Getting Started</h4>
                <p><strong>Grid Agent</strong> is an AI-powered interface for managing ThreeFold Grid resources. It combines natural language processing with direct Grid operations.</p>

                <h5>Initial Setup</h5>
                <ul>
                  <li><strong>Mnemonic Phrase:</strong> Your ThreeFold Grid wallet credentials</li>
                  <li><strong>Network:</strong> Choose mainnet for production, testnet for testing</li>
                  <li><strong>API Key:</strong> Get a free Gemini API key from Google AI Studio</li>
                </ul>
              </div>

              <!-- Using the Chat Interface -->
              <div class="docs-section">
                <h4>💬 Using the Chat Interface</h4>
                <p>Interact with your Grid resources using natural language commands.</p>

                <h5>Examples</h5>
                <div class="code-example">
                  <p>"Deploy a VM with 2 CPUs and 4GB RAM"</p>
                  <p>"Show me all my contracts"</p>
                  <p>"Deploy a wordpress instance"</p>
                </div>

                <h5>Understanding Responses</h5>
                <ul>
                  <li><strong>Steps:</strong> Each action is broken down into executable steps</li>
                  <li><strong>Commands:</strong> Actual CLI commands that will be executed</li>
                  <li><strong>Output:</strong> Real-time execution results</li>
                  <li><strong>Export:</strong> Save conversations for documentation</li>
                </ul>
              </div>

              <!-- Managing Personas -->
              <div class="docs-section">
                <h4>🎭 Managing Personas</h4>
                <p>Personas customize how the AI agent responds and behaves for different tasks.</p>

                <h5>How to Create and Use Personas</h5>
                <ol>
                  <li>Click <strong>"Personas"</strong> in the sidebar</li>
                  <li>Click the <strong>"+"</strong> button to create a new persona</li>
                  <li>Give it a descriptive name (e.g., "DevOps Engineer")</li>
                  <li>Add specific instructions for how the AI should behave</li>
                  <li>Click the persona card to activate it</li>
                </ol>

                <h5>When to Use Different Personas</h5>
                <ul>
                  <li><strong>Default:</strong> General Grid operations and basic deployments</li>
                  <li><strong>DevOps Engineer:</strong> Infrastructure automation and complex deployments</li>
                  <li><strong>System Administrator:</strong> Server management and troubleshooting</li>
                  <li><strong>Developer:</strong> Development environments and coding tasks</li>
                </ul>

                <h5>Example Persona Instructions</h5>
                <div class="code-example">
                  <p><strong>Name:</strong> DevOps Engineer</p>
                  <p><strong>Instructions:</strong> You are an experienced DevOps engineer. Always consider scalability, security, and automation. Prefer infrastructure as code approaches and suggest best practices for production deployments.</p>
                </div>
              </div>

              <!-- AI Configuration -->
              <div class="docs-section">
                <h4>🤖 AI Configuration</h4>
                <p>Configure the AI model and behavior settings.</p>

                <h5>How to Configure AI Settings</h5>
                <ol>
                  <li>Click <strong>"AI Configuration"</strong> in the sidebar</li>
                  <li><strong>Model Selection:</strong> Choose from available Gemini models (higher models = better responses but more tokens)</li>
                  <li><strong>API Key:</strong> Click "Show" to reveal, "Change" to update</li>
                  <li><strong>Export Summary:</strong> Enable AI-generated summaries when exporting conversations</li>
                </ol>

                <h5>Available Models</h5>
                <ul>
                  <li><strong>Gemini 3.0:</strong> Latest and most capable (recommended)</li>
                  <li><strong>Gemini 2.5:</strong> Good balance of capability and speed</li>
                  <li><strong>Gemini Flash:</strong> Fast responses, lower token usage</li>
                </ul>

                <h5>Export Options</h5>
                <ul>
                  <li><strong>With Summary:</strong> AI generates a summary of your conversation</li>
                  <li><strong>Without Summary:</strong> Saves tokens, just exports the raw conversation</li>
                </ul>
              </div>

              <!-- Grid Configuration -->
              <div class="docs-section">
                <h4>🌐 Grid Configuration</h4>
                <p>Set up your ThreeFold Grid connection and credentials.</p>

                <h5>How to Configure Grid Settings</h5>
                <ol>
                  <li>Click <strong>"Grid Configuration"</strong> in the sidebar</li>
                  <li><strong>Network:</strong> Choose mainnet for production, testnet for testing</li>
                  <li><strong>Mnemonic:</strong> Click "Show" to reveal your wallet phrase</li>
                  <li>Click <strong>"Save Configuration"</strong> when done</li>
                </ol>

                <h5>Network Options</h5>
                <ul>
                  <li><strong>Mainnet:</strong> Production network with real TFT tokens</li>
                  <li><strong>Testnet:</strong> Testing network with free tokens</li>
                  <li><strong>Devnet:</strong> Development network for advanced users</li>
                  <li><strong>QA:</strong> Quality assurance network</li>
                </ul>

                <h5>Mnemonic Security</h5>
                <ul>
                  <li>Your mnemonic is encrypted and stored locally</li>
                  <li>Never share your mnemonic with anyone</li>
                  <li>Use testnet first to practice before using mainnet</li>
                  <li>Keep backups of your mnemonic in a secure location</li>
                </ul>
              </div>

              <!-- Saving Conversations -->
              <div class="docs-section">
                <h4>💾 Saving Conversations</h4>
                <p>Export your AI conversations for documentation and reference.</p>

                <h5>How to Save Discussions</h5>
                <ol>
                  <li>Look for the <strong>download button</strong> (📥) in the top-right corner of the main screen</li>
                  <li>Click it to open the save dialog</li>
                  <li>Choose a filename (defaults to current date/time)</li>
                  <li>Select save location</li>
                  <li>Conversation saves as a Markdown file with full details</li>
                </ol>

                <h5>What Gets Saved</h5>
                <ul>
                  <li>All messages in the conversation</li>
                  <li>AI responses and reasoning steps</li>
                  <li>Commands executed and their outputs</li>
                  <li>Version information and active persona</li>
                  <li>Optional AI-generated summary (if enabled)</li>
                </ul>
              </div>

              <!-- Grid Operations -->
              <div class="docs-section">
                <h4>🌐 Grid Operations</h4>
                <p>Understanding ThreeFold Grid concepts and operations.</p>

                <h5>Key Concepts</h5>
                <ul>
                  <li><strong>Contracts:</strong> Agreements for resource usage</li>
                  <li><strong>Nodes:</strong> Physical servers providing resources</li>
                  <li><strong>Farms:</strong> Groups of nodes managed together</li>
                  <li><strong>Flists:</strong> Container images for deployments</li>
                </ul>

                <h5>Common Operations</h5>
                <ul>
                  <li><strong>VM Deployment:</strong> Virtual machines with custom specs</li>
                  <li><strong>Kubernetes:</strong> Container orchestration clusters</li>
                  <li><strong>Gateways:</strong> Load balancers and reverse proxies</li>
                  <li><strong>ZDB:</strong> Distributed database storage</li>
                </ul>
              </div>

              <!-- Troubleshooting -->
              <div class="docs-section">
                <h4>🔧 Troubleshooting</h4>

                <h5>Common Issues</h5>
                <ul>
                  <li><strong>"Agent not initialized":</strong> Check API key and network settings</li>
                  <li><strong>"Command failed":</strong> Verify network connectivity and credentials</li>
                  <li><strong>"No nodes available":</strong> Try different farm or node selection</li>
                </ul>

                <h5>Getting Help</h5>
                <ul>
                  <li><strong>ThreeFold Forum:</strong> Community support at forum.threefold.io</li>
                  <li><strong>Grid Manual:</strong> Comprehensive docs at manual.grid.tf</li>
                  <li><strong>GitHub Issues:</strong> Report bugs or request features at github.com/threefoldtech/grid-agent</li>
                </ul>
              </div>

              <!-- Tips -->
              <div class="docs-section">
                <h4>💡 Pro Tips</h4>
                <ul>
                  <li>Use specific resource requirements for better deployment success</li>
                  <li>Check contract status regularly to monitor resource usage</li>
                  <li>Export important conversations for documentation</li>
                  <li>Use personas for specialized workflows</li>
                  <li>Test on testnet before deploying to mainnet</li>
                </ul>
              </div>
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

  /* Documentation Styles */
  .docs-content {
    max-width: 800px;
    line-height: 1.6;
    text-align: left;
  }

  .docs-section {
    margin-bottom: 3rem;
    padding-bottom: 2rem;
    border-bottom: 1px solid var(--border);
  }

  .docs-section:last-child {
    border-bottom: none;
    margin-bottom: 0;
    padding-bottom: 0;
  }

  .docs-section h4 {
    margin: 0 0 1rem 0;
    font-size: 1.25rem;
    font-weight: 700;
    color: var(--text-primary);
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .docs-section h5 {
    margin: 1.5rem 0 0.75rem 0;
    font-size: 1rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .docs-section p {
    margin: 0 0 1rem 0;
    color: var(--text-primary);
  }

  .docs-section ul,
  .docs-section ol {
    margin: 0 0 1rem 0;
    padding-left: 1.5rem;
  }

  .docs-section li {
    margin-bottom: 0.5rem;
    color: var(--text-secondary);
  }

  .docs-section li strong {
    color: var(--text-primary);
    font-weight: 600;
  }

  .code-example {
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    padding: 1rem;
    margin: 1rem 0;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 0.9rem;
  }

  .code-example p {
    margin: 0.5rem 0;
    color: var(--accent);
    font-weight: 500;
  }

  .docs-section a {
    color: var(--accent);
    text-decoration: none;
  }

  .docs-section a:hover {
    text-decoration: underline;
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
