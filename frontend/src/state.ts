import type { Language } from '@bind/models';
import type { Info } from '@bind/provider';
import { reactive } from 'vue';

interface ProviderState {
    current: Info | null;
    providers: Info[];
}

interface LanguageState {
    detection: boolean;
    preferredSource: string;
    preferredTarget: string;
    sourceLanguages: Language[];
    targetLanguages: Language[];
    detecting: boolean;
    source: string;
    target: string;
    detected: string;
}

/**
 * The configured providers and the one in use. Null until the backend answered.
 */
export const providerState: ProviderState = reactive({
    current: null,
    providers: [],
});

/**
 * What the current provider offers (`languages:changed`) and what is chosen of it (`selection:changed`).
 * `detection` says the provider can detect, `detecting` that it does; `source` is empty while it does.
 */
export const languageState: LanguageState = reactive({
    detection: false,
    preferredSource: '',
    preferredTarget: '',
    sourceLanguages: [],
    targetLanguages: [],
    detecting: false,
    source: '',
    target: '',
    detected: '',
});

/**
 * The translation on screen and whether one is being made. `shown` stays false until the first one arrived.
 */
export const translationState = reactive({
    translation: '',
    shown: false,
    pending: false,
});

/**
 * Whether there is a stored translation to step to. Both stay false while the history is turned off.
 */
export const historyState = reactive({
    hasPrevious: false,
    hasNext: false,
});

/**
 * The failure that was reported last. It stays until the next one, so the error view can be reached again.
 */
export const errorState = reactive({
    message: '',
});
