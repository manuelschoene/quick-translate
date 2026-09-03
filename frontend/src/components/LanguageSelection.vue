<script lang="ts" setup>
import { route } from '@/router';
import type { Language } from '@/state';
import Button from '@comp/Button.vue';
import Panel from '@comp/Panel.vue';
import SearchBox from '@comp/SearchBox.vue';
import LayoutLabelled from '@lay/LayoutLabelled.vue';
import { cn } from '@lib/cn';
import { onKey } from '@lib/keyboard';
import { searchFilter } from '@lib/search';
import { computed, onMounted, ref, useTemplateRef, type DeepReadonly } from 'vue';

const props = defineProps<{
    currentTag: string;
    langs: DeepReadonly<Language[]>;
    pinned: readonly string[];
    label: string;
    changeFn: (tag: string) => void;
}>();

const search = ref('');

/**
 * The languages as they are listed.
 *
 * The pinned ones are only kept in front while nothing is typed: whoever searches is looking for a
 * certain language and not for the one they usually take, so from the first character on everything
 * is filtered and shown in the order it was handed over.
 */
const listedLanguages = computed(() => {
    if (search.value.length === 0) {
        return pinLanguages(props.langs, props.pinned);
    }

    return searchFilter(search.value, props.langs, (lang) => lang.name);
});

const list = useTemplateRef<HTMLElement>('list');

/**
 * Takes the language over and returns to the translation. Picking the one that is already set is not
 * caught here, the backend answers an unchanged combination from its cache.
 */
function selectLanguage(lang: Language): void {
    props.changeFn(lang.tag);
    route('translation');
}

/**
 * Takes the first language that is listed, which is what pressing enter in the search box means.
 */
function selectFirstLanguage(): void {
    const first = listedLanguages.value.at(0);
    if (first) selectLanguage(first);
}

/**
 * Puts the languages of the given tags in front of the rest and leaves the order of everything else
 * alone. A pinned language is moved and not copied, so it shows up once, and a tag that is not in
 * the list is skipped.
 */
function pinLanguages(languages: DeepReadonly<Language[]>, tags: readonly string[]): DeepReadonly<Language>[] {
    const pinned = tags
        .map((tag) => languages.find((language) => language.tag === tag))
        .filter((language) => language !== undefined);

    return [...pinned, ...languages.filter((language) => !pinned.includes(language))];
}

// The language that is set can sit anywhere in a list of some thirty entries, so it is scrolled to.
onMounted(() => list.value?.querySelector('[data-current]')?.scrollIntoView({ block: 'center' }));

onKey('Escape', () => {
    route('translation');
});
</script>

<template>
    <LayoutLabelled :back-action="() => route('translation')" :label="props.label">
        <Panel class="border-b" style="--wails-draggable: drag">
            <SearchBox
                @submit="selectFirstLanguage"
                placeholder="Search..."
                style="--wails-draggable: no-drag"
                title="Search for language"
                v-model="search"
            />
        </Panel>

        <Panel class="m-2 min-h-0 flex-1 overflow-hidden rounded-lg border border-white/10 p-0">
            <div
                class="grid h-full auto-rows-min grid-cols-3 content-start overflow-y-auto"
                ref="list"
                v-if="listedLanguages.length > 0"
            >
                <Button
                    :action="() => selectLanguage(lang)"
                    :class="
                        cn(
                            'min-h-(--control-height) py-2.5 hover:bg-linear-to-br hover:from-gray-500 hover:to-gray-600',
                            lang.tag === props.currentTag
                                && 'bg-linear-to-br from-accent-800 to-accent-900 hover:from-accent-700 hover:to-accent-800',
                        )
                    "
                    :data-current="lang.tag === props.currentTag ? '' : undefined"
                    :key="lang.tag"
                    :text="lang.name"
                    :title="lang.name"
                    class-text="text-center leading-tight"
                    inline
                    v-for="lang in listedLanguages"
                />
            </div>

            <div class="flex h-full items-center justify-center" v-else>
                <p class="text-gray-400">No languages found</p>
            </div>
        </Panel>
    </LayoutLabelled>
</template>
