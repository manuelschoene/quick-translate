import type { Language } from '@bind/models';
import { reactive } from 'vue';

interface ProviderState {
    current: string;
    providers: string[];
}

interface LanguageState {
    source: string;
    target: string;
    detectedSource: string;
    preferredSource: string;
    preferredTarget: string;
    detection: Language | null;
    sourceLanguages: Language[];
    targetLanguages: Language[];
}

/**
 * The translation services that are configured, held as slugs. What a slug is shown with belongs to
 * the frontend and is not part of this.
 */
export const providerState: ProviderState = reactive({
    current: '',
    providers: [],
});

/**
 * What the current provider offers and what is chosen of it. The lists change with the provider, the
 * chosen tags with every translation, and `detection` stays null for a provider that does not detect
 * the source language on its own.
 */
export const languageState: LanguageState = reactive({
    source: '',
    target: '',
    detectedSource: '',
    preferredSource: '',
    preferredTarget: '',
    detection: null,
    sourceLanguages: [],
    targetLanguages: [],
});

/**
 * What came back for the text that was translated last. Empty until the first translation was made.
 */
export const translationState = reactive({
    translation: '',
});

/**
 * Whether there is a stored translation to step to, in either direction. Both stay false while the
 * history is turned off.
 */
export const historyState = reactive({
    hasPrevious: false,
    hasNext: false,
});

/**
 * Whether the backend is working on what is shown. Not true for the calls beside the translation,
 * which leave the view as it is.
 */
export const requestState = reactive({
    pending: false,
});

/**
 * The failure that was reported last. It stays until the next one, so the error view can be left and
 * reached again without losing it.
 */
export const errorState = reactive({
    message: '',
});
