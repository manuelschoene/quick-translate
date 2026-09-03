import { languageState, type Language } from '@/state';
import { readonlyRefs } from '@lib/reactivity';
import { request } from '@services/request';
import { applyTranslation } from '@services/wire';
import { ChangeSource, ChangeTarget, SwitchLanguages } from '@wails/go/transport/Adapter';
import { computed, type DeepReadonly } from 'vue';

const {
    source,
    target,
    detectedSource,
    preferredSource,
    preferredTarget,
    detection,
    sourceLanguages,
    targetLanguages,
} = readonlyRefs(languageState);

/**
 * Every language that can be chosen as the source, sorted by name. Detection is part of it although
 * it is no language of the provider: to the user it is one more entry to pick from.
 */
const sourceOptions = computed(() =>
    sortByName(detection.value ? [detection.value, ...sourceLanguages.value] : sourceLanguages.value),
);

/**
 * Every language that can be chosen as the target, sorted by name.
 */
const targetOptions = computed(() => sortByName(targetLanguages.value));

/**
 * The source languages to keep within reach: detection first, because it is what most texts are
 * translated with, then the language the user configured.
 */
const pinnedSource = computed(() => tags(detection.value?.tag, preferredSource.value));

/**
 * The target language to keep within reach.
 */
const pinnedTarget = computed(() => tags(preferredTarget.value));

/**
 * The text the source language button shows: the name behind the tag, what detection made of the
 * last text in brackets, or a hint to pick one at all.
 */
const sourceLabel = computed(() => {
    if (source.value.length === 0) {
        return 'Not selected';
    }

    if (detection.value?.tag === source.value) {
        const detected = findByTag(detectedSource.value, sourceLanguages.value);

        return detected ? `${detection.value.name} (${detected.name})` : detection.value.name;
    }

    return findByTag(source.value, sourceLanguages.value)?.name ?? source.value;
});

/**
 * The text the target language button shows: the name behind the tag, or a hint to pick one at all.
 */
const targetLabel = computed(() => {
    if (target.value.length === 0) {
        return 'Not selected';
    }

    return findByTag(target.value, targetLanguages.value)?.name ?? target.value;
});

const exposed = {
    source,
    target,
    sourceLabel,
    targetLabel,
    sourceOptions,
    targetOptions,
    pinnedSource,
    pinnedTarget,
    changeSource,
    changeTarget,
    switchLanguages,
};

/**
 * Gives the components the languages of the current provider, the ones that are chosen, the text
 * they are shown with and the ways to change them. Every change translates the current text again,
 * so a component only picks a language and reads the new state.
 */
export function useLanguages(): typeof exposed {
    return exposed;
}

/**
 * Sets the source language and translates the current text again.
 */
async function changeSource(tag: string): Promise<void> {
    await request(() => ChangeSource(tag), applyTranslation);
}

/**
 * Sets the target language and translates the current text again.
 */
async function changeTarget(tag: string): Promise<void> {
    await request(() => ChangeTarget(tag), applyTranslation);
}

/**
 * Swaps source and target and translates the current text in the other direction.
 */
async function switchLanguages(): Promise<void> {
    await request(SwitchLanguages, applyTranslation);
}

/**
 * Collects the tags that are worth pinning and drops what is not set, so a provider without
 * detection and a user without a preference end up with fewer entries in front, not with empty ones.
 */
function tags(...candidates: (string | undefined)[]): string[] {
    return candidates.filter((tag): tag is string => tag !== undefined && tag.length > 0);
}

/**
 * Sorts a copy of the languages by name, which is what the user reads, not by tag.
 */
function sortByName(languages: readonly DeepReadonly<Language>[]): DeepReadonly<Language>[] {
    return languages.slice().sort((a, b) => a.name.localeCompare(b.name));
}

/**
 * Looks a language up by its tag. Returns nothing for a tag that is not in the list, which is the
 * regular case while the backend has not answered yet.
 */
function findByTag(tag: string, languages: readonly DeepReadonly<Language>[]): DeepReadonly<Language> | undefined {
    return languages.find((language) => language.tag === tag);
}
