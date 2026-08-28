<script lang="ts" setup>
import Button from '@comp/Button.vue';
import Panel from '@comp/Panel.vue';
import SearchBox from '@comp/SearchBox.vue';
import { showTranslation } from '@data/navigation';
import type { Language } from '@data/types';
import LayoutLabelled from '@lay/LayoutLabelled.vue';
import { cn } from '@utils/cn';
import { onKey } from '@utils/keyboard';
import { searchFilter } from '@utils/search';
import { computed, onMounted, ref, useTemplateRef, type DeepReadonly } from 'vue';

const props = defineProps<{
    currentTag: string;
    langs: DeepReadonly<Language[]>;
    label: string;
    changeFn: (tag: string) => void;
}>();

const sortedLanguages = props.langs.slice().sort((a, b) => a.name.localeCompare(b.name));

const search = ref<string>('');
const filteredLanguages = computed(() => searchFilter(search.value, sortedLanguages, (lang: Language) => lang.name));

const list = useTemplateRef<HTMLElement>('list');

const selectLanguage = (lang: Language) => {
    if (props.currentTag !== lang.tag) {
        props.changeFn(lang.tag);
    }

    showTranslation();
};

const selectFirstLanguage = () => {
    const first = filteredLanguages.value[0];
    if (first) selectLanguage(first);
};

onMounted(() => list.value?.querySelector('[data-current]')?.scrollIntoView({ block: 'center' }));

onKey('Escape', showTranslation);
</script>

<template>
    <LayoutLabelled :back-action="showTranslation" :label="label">
        <Panel class="border-b" style="--wails-draggable: drag">
            <SearchBox
                @submit="selectFirstLanguage"
                label="Search for language"
                placeholder="Search..."
                style="--wails-draggable: no-drag"
                v-model="search"
            />
        </Panel>

        <Panel class="m-2 min-h-0 flex-1 overflow-hidden rounded-lg border border-white/10 p-0">
            <div
                class="grid h-full auto-rows-min grid-cols-3 content-start overflow-y-auto"
                ref="list"
                v-if="filteredLanguages.length > 0"
            >
                <Button
                    :action="() => selectLanguage(lang)"
                    :class="
                        cn(
                            'min-h-(--control-height) py-2.5 hover:bg-linear-to-br hover:from-gray-500 hover:to-gray-600',
                            lang.tag === currentTag
                                && 'bg-linear-to-br from-accent-800 to-accent-900 hover:from-accent-700 hover:to-accent-800',
                        )
                    "
                    :data-current="lang.tag === currentTag ? '' : undefined"
                    :key="lang.tag"
                    :text="lang.name"
                    :title="lang.name"
                    class-text="text-center leading-tight"
                    inline
                    v-for="lang in filteredLanguages"
                />
            </div>

            <div class="flex h-full items-center justify-center" v-else>
                <p class="text-gray-400">No languages found</p>
            </div>
        </Panel>
    </LayoutLabelled>
</template>
