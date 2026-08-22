<script lang="ts" setup>
import { useLanguages } from '@use/useLanguages.ts';
import SelectionBaseView from '@views/SelectionBaseView.vue';
import { computed } from 'vue';
import { auto } from '@utils/language.ts';

const { source, changeSource, detection, sourceLanguages } = useLanguages();

const allLanguages = computed(() => {
    if (detection.value) {
        return [auto, ...sourceLanguages.value];
    }
    return sourceLanguages.value;
});
</script>

<template>
    <SelectionBaseView
        :change-fn="changeSource"
        :current-tag="source"
        :langs="allLanguages"
        label="Select Source Language"
    />
</template>
