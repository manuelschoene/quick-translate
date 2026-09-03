<script lang="ts" setup>
import { useLanguages } from '@use/useLanguages';
import SelectionBaseView from '@views/SelectionBaseView.vue';

const { source, sourceOptions, pinnedSource, changeSource } = useLanguages();
</script>

<template>
    <SelectionBaseView
        :change-fn="changeSource"
        :current-tag="source"
        :langs="sourceOptions"
        :pinned="pinnedSource"
        label="Select Source Language"
    />
</template>
