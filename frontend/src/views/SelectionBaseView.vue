<script lang="ts" setup>
import { showTranslation } from '@data/navigation';
import LayoutLabelled from '@lay/LayoutLabelled.vue';
import Panel from '@comp/Panel.vue';
import SearchBox from '@comp/SearchBox.vue';
import { computed, DeepReadonly, ref } from 'vue';
import { Language } from '@data/types';
import { searchFilter } from '@utils/search';

const props = defineProps<{
    currentTag: string;
    langs: DeepReadonly<Language[]>;
    label: string;
    changeFn: (tag: string) => void;
}>();

const sortedLanguages = props.langs.slice().sort((a, b) => a.name.localeCompare(b.name));

const search = ref<string>('');
const filteredLanguages = computed(() => searchFilter(search.value, sortedLanguages, (lang: Language) => lang.name));

const selectLanguage = (lang: Language) => {
    if (props.currentTag == lang.tag) return;
    props.changeFn(lang.tag);
    showTranslation();
};
</script>x

<template>
    <LayoutLabelled :back-action="showTranslation" :label="label">
        <Panel class="border-b">
            <SearchBox label="Search for language" placeholder="Search..." v-model="search" />
        </Panel>

        <Panel class="m-2 min-h-0 flex-1 overflow-hidden rounded-lg border border-white/10 p-0">
            <div class="grid h-full grid-cols-3 gap-2 overflow-y-auto" v-if="filteredLanguages.length > 0">
                <button
                    class="group h-fit p-2 text-sm hover:cursor-pointer hover:bg-linear-to-br hover:from-gray-500 hover:to-gray-600 hover:font-bold focus-visible:ring-2 focus-visible:ring-gray-400 focus-visible:outline-none focus-visible:ring-inset"
                    :key="lang.tag"
                    @click="selectLanguage(lang)"
                    v-for="lang in filteredLanguages"
                >
                    {{ lang.name }}
                </button>
            </div>

            <div class="flex h-full items-center justify-center" v-else>
                <p class="text-gray-400">No languages found</p>
            </div>
        </Panel>
    </LayoutLabelled>
</template>
