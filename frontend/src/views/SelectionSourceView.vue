<script lang="ts" setup>
import LanguageSelection from '@comp/LanguageSelection.vue';
import { useLanguages } from '@use/useLanguages';

const { source, sourceOptions, pinnedSource, changeSource } = useLanguages();
</script>

<template>
    <LanguageSelection
        :change-fn="changeSource"
        :current-tag="source"
        :langs="sourceOptions"
        :pinned="pinnedSource"
        label="Select Source Language"
    />
</template>
