<script lang="ts" setup>
import { useLanguages } from '@use/useLanguages';
import SelectionBaseView from '@views/SelectionBaseView.vue';

const { target, targetOptions, pinnedTarget, changeTarget } = useLanguages();
</script>

<template>
    <SelectionBaseView
        :change-fn="changeTarget"
        :current-tag="target"
        :langs="targetOptions"
        :pinned="pinnedTarget"
        label="Select Target Language"
    />
</template>
