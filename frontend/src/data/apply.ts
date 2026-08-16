import { historyState, languageState, providerState, translationState } from '@data/state';
import type { Language } from '@data/types';
import type { models, transport } from '@wails/go/models';

/**
 * Takes over the configured providers together with the one that is in use.
 */
export function applyProviders(dto: transport.ProviderDto): void {
    providerState.current = dto.Current;
    providerState.providers = dto.Providers ?? [];
}

/**
 * Takes over the languages of the current provider together with the ones that are chosen. Sent
 * whenever the provider changes, because the languages that can be chosen belong to it.
 */
export function applyLanguages(dto: transport.LanguageDto): void {
    languageState.source = dto.Source;
    languageState.target = dto.Target;
    languageState.detection = dto.Detection;
    languageState.sourceLanguages = toLanguages(dto.SourceLanguages);
    languageState.targetLanguages = toLanguages(dto.TargetLanguages);
}

/**
 * Takes over the answer of a translation, which reaches further than the translated text alone. It
 * also carries the languages, because the core adjusts one of them when it collides with the other,
 * and the navigation flags, because a translation that was made is one more step to go back to.
 */
export function applyTranslation(dto: transport.TranslationDto): void {
    languageState.source = dto.Source;
    languageState.target = dto.Target;
    languageState.detectedSource = dto.DetectedSource;

    translationState.text = dto.Text;
    translationState.translation = dto.Translation;

    historyState.hasPrevious = dto.HasPrevious;
    historyState.hasNext = dto.HasNext;
}

/**
 * Takes over the whole state, which is sent when the provider changes and when stepping through the
 * history, as a stored translation brings its own provider and languages with it.
 */
export function applyFull(dto: transport.FullDto): void {
    providerState.current = dto.Provider;

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
    return (languages ?? []).map(toLanguage);
}

/**
 * Maps a language of the backend to the shape the frontend works with. Together with the functions
 * above this is the only place in the project that knows the Go field names.
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
