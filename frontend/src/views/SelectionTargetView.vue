<script lang="ts" setup>
import { useLanguages } from '@use/useLanguages.ts';
import SelectionBaseView from '@views/SelectionBaseView.vue';

const { target, changeTarget, targetLanguages } = useLanguages();
</script>

<template>
    <SelectionBaseView
        :change-fn="changeTarget"
        :current-tag="target"
        :langs="targetLanguages"
        label="Select Target Language"
    />
</template>
