import { reactive } from 'vue';

/**
 * The state of the whole application, split by what it describes and held once per module. A single
 * window works on a single core, so there is nothing to keep apart per component.
 *
 * The state is written by `services/wire.ts`, which unpacks what the backend answers, and by
 * `services/report.ts` and `services/request.ts`, which note how the last call went. It is read
 * through the composables, which hand it out read only, so it can only be changed by making a call.
 */

/**
 * A language a provider offers, in the shape the frontend works with. The backend delivers the same
 * information with Go field names, which `services/wire.ts` maps away.
 */
export interface Language {
    tag: string;
    name: string;
    source: boolean;
    target: boolean;
    stable: boolean;
}

/**
 * The translation services that are configured, held as slugs: what a slug is shown with belongs to
 * the frontend and lives in `composables/useProviders.ts`.
 */
interface ProviderState {
    current: string;
    providers: string[];
}

/**
 * What the current provider offers and what is chosen of it. The lists change with the provider, the
 * chosen tags change with every translation, and `detection` is missing for a provider that does not
 * detect the source language on its own.
 */
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

export const providerState: ProviderState = reactive({
    current: '',
    providers: [],
});

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

export const translationState = reactive({
    translation: '',
});

export const historyState = reactive({
    hasPrevious: false,
    hasNext: false,
});

export const requestState = reactive({
    pending: false,
});

export const errorState = reactive({
    message: '',
});
