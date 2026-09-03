import type { Language } from '@/state';
import { historyState, languageState, providerState, translationState } from '@/state';
import type { models, transport } from '@wails/go/models';

/**
 * The only place in the project that knows the wire format. It takes the DTOs apart and hands each
 * part to the state group it belongs to, which is why it stays one module although the state is
 * split by domain: a `TranslationDto` carries the chosen languages, the translated text and the
 * history flags at once, so whoever unpacks it has to reach into three groups anyway.
 */

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
 * Takes over what the current provider offers. Sent whenever the provider changes, because the
 * languages that can be chosen belong to it. Which of them are chosen is not part of this, it
 * arrives with the translation.
 *
 * The lists come in no promised order, so whoever shows them sorts them. Detection is missing for a
 * provider that does not detect the source language on its own.
 */
export function applyLanguages(dto: transport.LanguageDto): void {
    languageState.preferredSource = dto.PreferredSource;
    languageState.preferredTarget = dto.PreferredTarget;
    languageState.detection = dto.Detection ? toLanguage(dto.Detection) : null;
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

    translationState.translation = dto.Translation;

    historyState.hasPrevious = dto.HasPrevious;
    historyState.hasNext = dto.HasNext;
}

/**
 * Takes over the whole state, which is sent for the first render, when the provider changes and when
 * stepping through the history, as a stored translation brings its own provider and languages with
 * it.
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
 * Maps a language of the backend to the shape the frontend works with.
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
