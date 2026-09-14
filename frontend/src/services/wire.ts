import { historyState, languageState, providerState, translationState } from '@/state';
import type { Language } from '@bind/models';
import type { FullDto, LanguageDto, ProviderDto, TranslationDto } from '@bind/transport';

/**
 * Takes over the configured providers together with the one that is in use.
 */
export function applyProviders(dto: ProviderDto): void {
    providerState.current = dto.Current;
    providerState.providers = dto.Providers;
}

/**
 * Takes over what the current provider offers. Which of the languages are chosen is not part of it,
 * that arrives with the translation.
 *
 * The lists come in no promised order, so whoever shows them sorts them.
 */
export function applyLanguages(dto: LanguageDto): void {
    languageState.preferredSource = dto.PreferredSource;
    languageState.preferredTarget = dto.PreferredTarget;
    languageState.detection = dto.Detection;
    languageState.sourceLanguages = present(dto.SourceLanguages);
    languageState.targetLanguages = present(dto.TargetLanguages);
}

/**
 * Takes over the answer of a translation, which reaches further than the translated text alone: the
 * backend adjusts a language when it collides with the other, and a translation that was made is one
 * more step to go back to.
 *
 * Accepts nothing and keeps the state as it is then, see `applyFull`.
 */
export function applyTranslation(dto: TranslationDto | null): void {
    if (!dto) {
        return;
    }

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
 *
 * Accepts nothing because the backend answers with a pointer, which the binding generator types as
 * nullable. The backend only ever sends nil together with an error, and that rejects the call before
 * it gets here, so nothing is simply left as it was.
 */
export function applyFull(dto: FullDto | null): void {
    if (!dto) {
        return;
    }

    if (dto.Provider) {
        applyProviders(dto.Provider);
    }

    if (dto.Languages) {
        applyLanguages(dto.Languages);
    }

    applyTranslation(dto.Translation);
}

/**
 * Drops the empty entries of a language list. The backend holds its languages as pointers, so the
 * binding generator types every entry as nullable, although a list never carries a nil one.
 */
function present(languages: (Language | null)[]): Language[] {
    return languages.filter((language): language is Language => language !== null);
}
