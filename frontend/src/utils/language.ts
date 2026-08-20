import type { Language } from '@data/types';
import { useLanguages } from '@use/useLanguages';
import { computed, type ComputedRef, type DeepReadonly } from 'vue';

/**
 * Gives the text the source language button shows. The tag alone is user-unfriendly, so the name
 * of the language is looked up. If the source language is set to auto, the detected language is
 * shown in brackets. If no source language is set, a hint is shown for the user to set the source
 * language.
 */
export function displaySourceLanguage(): ComputedRef<string> {
    const { source, detectedSource, sourceLanguages } = useLanguages();

    return computed(() => {
        if (source.value.length === 0) {
            return 'Not selected';
        }

        if (source.value === 'auto') {
            if (detectedSource.value.length > 0) {
                const detectedLanguage = findLanguageByTag(detectedSource.value, sourceLanguages.value);
                if (detectedLanguage) {
                    return `Auto (${detectedLanguage.name})`;
                }
            }

            return 'Auto';
        }

        const language = findLanguageByTag(source.value, sourceLanguages.value);

        return language ? language.name : source.value;
    });
}

/**
 * Gives the text the target language button shows. The tag alone is user-unfriendly, so the name
 * of the language is looked up. If no target language is set, a hint is shown for the user to set
 * the target language.
 */
export function displayTargetLanguage(): ComputedRef<string> {
    const { target, targetLanguages } = useLanguages();

    return computed(() => {
        if (target.value.length === 0) {
            return 'Not selected';
        }

        const language = findLanguageByTag(target.value, targetLanguages.value);

        return language ? language.name : target.value;
    });
}

/**
 * Looks a language up by its tag. The list arrives as the read only view the composables hand out,
 * so the language that is found is read only as well.
 */
function findLanguageByTag(tag: string, languages: DeepReadonly<Language[]>): DeepReadonly<Language> | undefined {
    return languages.find((language) => language.tag === tag);
}
