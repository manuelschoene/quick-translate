import { languageState } from '@/state';
import type { Options as OptionsModel, Selection as SelectionModel } from '@bind/language';
import { Detect, Options, Selection, SetSource, SetTarget, Switch } from '@bind/language/service';
import { Language } from '@bind/models';
import { LanguagesChanged, SelectionChanged } from '@events';
import { on } from '@lib/events';
import { readonlyRefs } from '@lib/reactivity';
import { request } from '@services/request';
import { retranslate } from '@use/useTranslation';
import { computed, type DeepReadonly } from 'vue';

/**
 * The tag the detection entry is listed with. Detection is no language of the provider, only a flag, so it gets
 * a tag of its own that no BCP 47 tag can be.
 */
const detectionTag = '@detection';

/**
 * The entry that turns detection on, listed among the source languages.
 */
const detectionEntry = new Language({ tag: detectionTag, name: 'Auto', source: true, stable: true });

const {
    detection,
    preferredSource,
    preferredTarget,
    sourceLanguages,
    targetLanguages,
    detecting,
    source: selectedSource,
    target,
    detected,
} = readonlyRefs(languageState);

/**
 * The tag of the chosen source, the detection entry while detecting.
 */
const source = computed(() => (detecting.value ? detectionTag : selectedSource.value));

/**
 * Every language that can be chosen as the source, sorted by name, with the detection entry if the provider
 * detects.
 */
const sourceOptions = computed(() =>
    sortByName(detection.value ? [detectionEntry, ...sourceLanguages.value] : sourceLanguages.value),
);

/**
 * Every language that can be chosen as the target, sorted by name.
 */
const targetOptions = computed(() => sortByName(targetLanguages.value));

/**
 * The source languages to keep within reach: detection first, then the configured one.
 */
const pinnedSource = computed(() => tags(detection.value ? detectionTag : undefined, preferredSource.value));

/**
 * The target language to keep within reach.
 */
const pinnedTarget = computed(() => tags(preferredTarget.value));

/**
 * The text of the source button: the language name, detection with what it detected last in brackets, or a hint.
 */
const sourceLabel = computed(() => {
    if (detecting.value) {
        const found = findByTag(detected.value, sourceLanguages.value);

        return found ? `${detectionEntry.name} (${found.name})` : detectionEntry.name;
    }

    if (selectedSource.value.length === 0) {
        return 'Not selected';
    }

    return findByTag(selectedSource.value, sourceLanguages.value)?.name ?? selectedSource.value;
});

/**
 * The text of the target button: the language name or a hint.
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
 * Gives the components the languages of the current provider, the chosen ones, their labels and the ways to
 * change them. Every change translates the shown text again.
 */
export function useLanguages(): typeof exposed {
    return exposed;
}

/**
 * Listens to the language lists and the selection. The lists always arrive first, so a selected tag finds its
 * name. Returns the unsubscribe.
 */
export function listen(): () => void {
    const offs = [on(LanguagesChanged, applyOptions), on(SelectionChanged, applySelection)];

    return () => {
        offs.forEach((off) => {
            off();
        });
    };
}

/**
 * Fetches the lists and the selection, for the first render.
 */
export async function load(): Promise<void> {
    const [options, selection] = await Promise.all([Options(), Selection()]);

    applyOptions(options);
    applySelection(selection);
}

/**
 * Sets the source language, or turns detection on for the detection entry, and translates again.
 */
async function changeSource(tag: string): Promise<void> {
    await request(async () => {
        await (tag === detectionTag ? Detect() : SetSource(tag));
        await retranslate();
    });
}

/**
 * Sets the target language and translates again.
 */
async function changeTarget(tag: string): Promise<void> {
    await request(async () => {
        await SetTarget(tag);
        await retranslate();
    });
}

/**
 * Swaps source and target and translates in the other direction.
 */
async function switchLanguages(): Promise<void> {
    await request(async () => {
        await Switch();
        await retranslate();
    });
}

/**
 * Takes over what the provider offers.
 */
function applyOptions(options: OptionsModel): void {
    languageState.detection = options.detection;
    languageState.preferredSource = options.preferredSource;
    languageState.preferredTarget = options.preferredTarget;
    languageState.sourceLanguages = options.sourceLanguages;
    languageState.targetLanguages = options.targetLanguages;
}

/**
 * Takes over what is chosen.
 */
function applySelection(selection: SelectionModel): void {
    languageState.detecting = selection.detection;
    languageState.source = selection.source;
    languageState.target = selection.target;
    languageState.detected = selection.detected;
}

/**
 * Collects the set tags, so a missing preference leaves fewer pinned entries instead of empty ones.
 */
function tags(...candidates: (string | undefined)[]): string[] {
    return candidates.filter((tag): tag is string => tag !== undefined && tag.length > 0);
}

/**
 * Sorts a copy of the languages by name.
 */
function sortByName(languages: readonly DeepReadonly<Language>[]): DeepReadonly<Language>[] {
    return languages.slice().sort((a, b) => a.name.localeCompare(b.name));
}

/**
 * Looks a language up by its tag. Nothing for an unknown tag, the regular case before the backend answered.
 */
function findByTag(tag: string, languages: readonly DeepReadonly<Language>[]): DeepReadonly<Language> | undefined {
    return languages.find((language) => language.tag === tag);
}
