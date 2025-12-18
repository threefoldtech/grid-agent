import { writable } from 'svelte/store';
import { ActivateProfile, DeactivateProfile } from "../../wailsjs/go/main/App.js"; // Adjust path if needed

export interface Profile {
    id: string;
    name: string;
    instructions: string;
}

export interface Settings {
    mnemonics: string;
    network: string;
    geminiApiKey: string;
    model: string;
    theme: string;
    isConfigured: boolean;
    profiles: Profile[];
    activeProfileID: string;
    enableExportSummary: boolean;
}

export const themeStore = writable('dark');
export const settingsStore = writable<Settings>({
    mnemonics: '',
    network: 'main',
    geminiApiKey: '',
    model: '',
    theme: 'dark',
    isConfigured: false,
    profiles: [],
    activeProfileID: '',
    enableExportSummary: false
});
export const messagesStore = writable([]);

// Helper actions to centralize profile logic
export async function activateProfile(id: string) {
    try {
        const newSettings = await ActivateProfile(id);
        settingsStore.set(newSettings);
    } catch (err) {
        throw err; // Re-throw to let component handle specific UI feedback if needed
    }
}

export async function deactivateProfile() {
    try {
        const newSettings = await DeactivateProfile();
        settingsStore.set(newSettings);
    } catch (err) {
        throw err;
    }
}
