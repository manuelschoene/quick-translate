import type { Language } from '@/state';
import { historyState, languageState, providerState, translationState } from '@/state';
import type { models, transport } from '@wails/go/models';

/**
 * Takes over the configured providers together with the one that is in use.
 */
export function applyProviders(dto: transport.ProviderDto): void {
    providerState.current = dto.Current;
    // Wails types a Go slice as an array, but an empty one arrives as null.
    // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
    providerState.providers = dto.Providers ?? [];
}

/**
 * Takes over what the current provider offers. Which of the languages are chosen is not part of it,
 * that arrives with the translation.
 *
 * The lists come in no promised order, so whoever shows them sorts them.
 */
export function applyLanguages(dto: transport.LanguageDto): void {
    languageState.preferredSource = dto.PreferredSource;
    languageState.preferredTarget = dto.PreferredTarget;
    languageState.detection = dto.Detection ? toLanguage(dto.Detection) : null;
    languageState.sourceLanguages = toLanguages(dto.SourceLanguages);
    languageState.targetLanguages = toLanguages(dto.TargetLanguages);
}

/**
 * Takes over the answer of a translation, which reaches further than the translated text alone: the
 * backend adjusts a language when it collides with the other, and a translation that was made is one
 * more step to go back to.
 */
export function applyTranslation(dto: transport.TranslationDto): void {
    languageState.source = dto.Source;
    languageState.target = dto.Target;
    languageState.detectedSource = dto.DetectedSource;

    translationState.translation = dto.Translation;

    historyState.hasPrevious = dto.HasPrevious;
    historyState.hasNext = dto.HasNext;
}

/**
 * Takes over everything at once, which is what the first render and every step through the history
 * need: a stored translation brings its own provider and languages with it.
 */
export function applyFull(dto: transport.FullDto): void {
    if (dto.Provider) {
        applyProviders(dto.Provider);
    }

    if (dto.Languages) {
        applyLanguages(dto.Languages);
    }

    if (dto.Translation) {
        applyTranslation(dto.Translation);
    }
}

/**
 * Maps a list of languages of the backend to the shape the frontend works with. An empty list
 * arrives as null and not as an empty array, because that is what an empty Go slice turns into.
 */
function toLanguages(languages: models.Language[]): Language[] {
    // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
    return (languages ?? []).map(toLanguage);
}

/**
 * Maps a language of the backend to the shape the frontend works with. Together with the functions
 * above, this file is the only one that knows the Go field names.
 */
function toLanguage(language: models.Language): Language {
    return {
        tag: language.Tag,
        name: language.Name,
        source: language.Source,
        target: language.Target,
        stable: language.Stable,
    };
}
