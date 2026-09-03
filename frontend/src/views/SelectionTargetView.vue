<script lang="ts" setup>
import LanguageSelection from '@comp/LanguageSelection.vue';
import { useLanguages } from '@use/useLanguages';

const { target, targetOptions, pinnedTarget, changeTarget } = useLanguages();
</script>

<template>
    <LanguageSelection
        :change-fn="changeTarget"
        :current-tag="target"
        :langs="targetOptions"
        :pinned="pinnedTarget"
        label="Select Target Language"
    />
</template>
