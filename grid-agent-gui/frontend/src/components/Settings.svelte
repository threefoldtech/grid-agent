<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { fade, fly } from "svelte/transition";
  import {
    AddProfile,
    UpdateProfile,
    DeleteProfile,
    ActivateProfile,
    DeactivateProfile,
  } from "../../wailsjs/go/main/App.js";
  import { settingsStore } from "../stores/stores";

  const dispatch = createEventDispatcher();

  export let show = false;

  let profiles = [];
  let activeProfileID = "";
  let editingProfile = null; // null means list view, object means editing
  let isCreating = false;

  // Form fields
  let formName = "";
  let formInstructions = "";
  let error = "";
  let showDeleteModal = false;
  let profileToDelete = null;

  $: {
    if ($settingsStore) {
      profiles = $settingsStore.profiles || [];
      activeProfileID = $settingsStore.activeProfileID || "";
    }
  }

  function close() {
    dispatch("close");
  }

  function startCreate() {
    isCreating = true;
    editingProfile = null;
    formName = "";
    formInstructions = "";
    error = "";
  }

  function startEdit(profile) {
    isCreating = false;
    editingProfile = profile;
    formName = profile.name;
    formInstructions = profile.instructions;
    error = "";
  }

  function cancelEdit() {
    editingProfile = null;
    isCreating = false;
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
    } catch (e) {
      error = e.message;
    }
  }

  function deleteProfile(profile) {
    profileToDelete = profile;
    showDeleteModal = true;
  }

  function cancelDelete() {
    showDeleteModal = false;
    profileToDelete = null;
  }

  async function confirmDelete() {
    if (!profileToDelete) return;
    showDeleteModal = false;
    try {
      const newSettings = await DeleteProfile(profileToDelete.id);
      settingsStore.set(newSettings);
    } catch (e) {
      error = e.message;
    }
    profileToDelete = null;
  }

  async function toggleActive(id) {
    try {
      let newSettings;
      if (activeProfileID === id) {
        newSettings = await DeactivateProfile();
      } else {
        newSettings = await ActivateProfile(id);
      }
      settingsStore.set(newSettings);
    } catch (e) {
      error = e.message;
    }
  }
</script>

{#if show}
  <div class="modal-overlay" on:click={close} transition:fade>
    <div class="modal" on:click|stopPropagation transition:fly={{ y: 20 }}>
      <div class="header">
        <h2>Personalization Settings</h2>
        <button class="close-btn" on:click={close}>&times;</button>
      </div>

      <div class="content">
        {#if isCreating || editingProfile}
          <div class="form" in:fade>
            <h3>{isCreating ? "Create New Profile" : "Edit Profile"}</h3>

            <div class="form-group">
              <label for="name">Profile Name</label>
              <input
                type="text"
                id="name"
                bind:value={formName}
                placeholder="e.g., Pirate Mode, Python Expert"
              />
            </div>

            <div class="form-group">
              <label for="instructions">Custom Instructions</label>
              <textarea
                id="instructions"
                bind:value={formInstructions}
                placeholder="Enter instructions for the agent (e.g., 'Always speak like a pirate', 'Focus on concise code examples')..."
                rows="5"
              ></textarea>
              <p class="hint">
                These instructions will be added to the system prompt.
              </p>
            </div>

            {#if error}
              <div class="error">{error}</div>
            {/if}

            <div class="actions">
              <button class="btn secondary" on:click={cancelEdit}>Cancel</button
              >
              <button class="btn primary" on:click={saveProfile}>Save</button>
            </div>
          </div>
        {:else}
          <div class="profile-list" in:fade>
            <div class="list-header">
              <h3>Your Profiles</h3>
              <button class="btn primary small" on:click={startCreate}
                >+ New Profile</button
              >
            </div>

            {#if profiles.length === 0}
              <div class="empty-state">
                <p>No profiles created yet.</p>
                <p class="sub">
                  Create a profile to customize the agent's behavior.
                </p>
              </div>
            {:else}
              <div class="profiles">
                {#each profiles as profile (profile.id)}
                  <div
                    class="profile-item"
                    class:active={activeProfileID === profile.id}
                  >
                    <div class="profile-info">
                      <div class="profile-name">
                        {profile.name}
                        {#if activeProfileID === profile.id}
                          <span class="badge">Active</span>
                        {/if}
                      </div>
                      <div class="profile-preview">
                        {profile.instructions.slice(0, 50)}{profile.instructions
                          .length > 50
                          ? "..."
                          : ""}
                      </div>
                    </div>
                    <div class="profile-actions">
                      <button
                        class="btn-icon"
                        class:active={activeProfileID === profile.id}
                        title={activeProfileID === profile.id
                          ? "Deactivate"
                          : "Activate"}
                        on:click={() => toggleActive(profile.id)}
                      >
                        {#if activeProfileID === profile.id}
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
                            ><rect
                              width="18"
                              height="18"
                              x="3"
                              y="3"
                              rx="2"
                            /></svg
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
                            ><polygon points="6 3 20 12 6 21 6 3" /></svg
                          >
                        {/if}
                      </button>
                      <button
                        class="btn-icon"
                        title="Edit"
                        on:click={() => startEdit(profile)}
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
                          ><path
                            d="M21.174 6.812a1 1 0 0 0-3.986-3.987L3.842 16.174a2 2 0 0 0-.5.83l-1.321 4.352a.5.5 0 0 0 .623.622l4.353-1.32a2 2 0 0 0 .83-.497z"
                          /><path d="m15 5 4 4" /></svg
                        >
                      </button>
                      <button
                        class="btn-icon delete"
                        title="Delete"
                        on:click={() => deleteProfile(profile)}
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
                          ><path d="M3 6h18" /><path
                            d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"
                          /><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" /><line
                            x1="10"
                            x2="10"
                            y1="11"
                            y2="17"
                          /><line x1="14" x2="14" y1="11" y2="17" /></svg
                        >
                      </button>
                    </div>
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<!-- Delete Confirmation Modal -->
{#if showDeleteModal}
  <div class="modal-overlay" on:click={cancelDelete} transition:fade>
    <div class="modal delete-modal" on:click|stopPropagation transition:fade>
      <h2>Confirm Delete</h2>
      <p>
        Are you sure you want to delete the profile "{profileToDelete?.name}"?
      </p>
      <p class="warning">This action cannot be undone.</p>
      <div class="modal-actions">
        <button class="btn secondary" on:click={cancelDelete}>Cancel</button>
        <button class="btn danger" on:click={confirmDelete}>Delete</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    backdrop-filter: blur(4px);
  }

  .modal {
    background: var(--bg-secondary);
    border-radius: 1rem;
    width: 90%;
    max-width: 600px;
    max-height: 85vh;
    display: flex;
    flex-direction: column;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
    border: 1px solid var(--border);
  }

  .header {
    padding: 1.5rem;
    border-bottom: 1px solid var(--border);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .header h2 {
    margin: 0;
    font-size: 1.25rem;
    color: var(--text-primary);
  }

  .close-btn {
    background: none;
    border: none;
    color: var(--text-secondary);
    font-size: 1.5rem;
    cursor: pointer;
    padding: 0;
    line-height: 1;
  }

  .close-btn:hover {
    color: var(--text-primary);
  }

  .content {
    padding: 1.5rem;
    overflow-y: auto;
  }

  /* List View */
  .list-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
  }

  .list-header h3 {
    margin: 0;
    font-size: 1rem;
    color: var(--text-secondary);
  }

  .profiles {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .profile-item {
    background: var(--bg-primary);
    border: 1px solid var(--border);
    border-radius: 0.75rem;
    padding: 1rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
    transition: all 0.2s;
  }

  .profile-item.active {
    border-color: var(--accent);
    background: rgba(59, 130, 246, 0.05);
  }

  .profile-info {
    flex: 1;
  }

  .profile-name {
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: 0.25rem;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .badge {
    background: var(--accent);
    color: white;
    font-size: 0.7rem;
    padding: 0.1rem 0.4rem;
    border-radius: 1rem;
    text-transform: uppercase;
    font-weight: 700;
  }

  .profile-preview {
    font-size: 0.875rem;
    color: var(--text-secondary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 300px;
  }

  .profile-actions {
    display: flex;
    gap: 0.5rem;
  }

  .btn-icon {
    background: var(--bg-tertiary);
    border: none;
    width: 32px;
    height: 32px;
    border-radius: 0.5rem;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.2s ease;
    font-size: 1rem;
    color: var(--text-secondary);
  }

  .btn-icon:hover {
    background: var(--border);
    color: var(--text-primary);
    transform: translateY(-1px);
  }

  .btn-icon:active {
    transform: translateY(0);
  }

  .btn-icon.active {
    color: var(--accent);
    background: rgba(59, 130, 246, 0.1);
  }

  .btn-icon.active:hover {
    background: rgba(59, 130, 246, 0.15);
  }

  .btn-icon.delete:hover {
    background: rgba(239, 68, 68, 0.15);
    color: var(--error);
  }

  .empty-state {
    text-align: center;
    padding: 3rem 1rem;
    color: var(--text-secondary);
    background: var(--bg-primary);
    border-radius: 0.75rem;
    border: 1px dashed var(--border);
  }

  .empty-state .sub {
    font-size: 0.875rem;
    margin-top: 0.5rem;
    opacity: 0.7;
  }

  /* Form View */
  .form h3 {
    margin: 0 0 1.5rem 0;
    color: var(--text-primary);
  }

  .form-group {
    margin-bottom: 1.5rem;
  }

  label {
    display: block;
    margin-bottom: 0.5rem;
    color: var(--text-secondary);
    font-size: 0.875rem;
  }

  input,
  textarea {
    width: 100%;
    padding: 0.75rem;
    border-radius: 0.5rem;
    border: 1px solid var(--border);
    background: var(--bg-primary);
    color: var(--text-primary);
    font-size: 1rem;
    font-family: inherit;
  }

  textarea {
    resize: vertical;
  }

  input:focus,
  textarea:focus {
    outline: none;
    border-color: var(--accent);
  }

  .hint {
    font-size: 0.75rem;
    color: var(--text-secondary);
    margin-top: 0.5rem;
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 1rem;
    margin-top: 2rem;
  }

  .error {
    color: var(--error);
    background: rgba(239, 68, 68, 0.1);
    padding: 0.75rem;
    border-radius: 0.5rem;
    margin-bottom: 1rem;
    font-size: 0.875rem;
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

  .btn.small {
    padding: 0.4rem 0.8rem;
    font-size: 0.8rem;
  }

  .btn.primary {
    background: var(--accent);
    color: white;
  }

  .btn.primary:hover {
    background: var(--accent-hover);
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

  /* Delete Modal */
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

  /* Delete confirmation modal specific styling */
  .delete-modal {
    padding: 2rem;
    max-width: 400px;
    max-height: none;
  }
</style>
