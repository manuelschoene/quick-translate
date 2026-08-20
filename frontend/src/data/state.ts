import type { Language, ViewName } from '@data/types';
import { reactive } from 'vue';

/**
 * The state of the whole application, split by what it describes and held once per module. A single
 * window works on a single core, so there is nothing to keep apart per component.
 *
 * The state is written by `apply.ts`, `navigation.ts` and `report.ts` and read through the
 * composables, which hand it out read only.
 */

export const providerState = reactive({
    current: '',
    providers: [] as string[],
});

export const languageState = reactive({
    source: '',
    target: '',
    detectedSource: '',
    detection: false,
    sourceLanguages: [] as Language[],
    targetLanguages: [] as Language[],
});

export const translationState = reactive({
    text: '',
    translation: '',
});

export const historyState = reactive({
    hasPrevious: false,
    hasNext: false,
});

export const requestState = reactive({
    pending: false,
});

export const viewState = reactive({
    current: 'translation' as ViewName,
});

export const errorState = reactive({
    message: '',
});
