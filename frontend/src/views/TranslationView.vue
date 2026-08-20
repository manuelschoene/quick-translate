<script lang="ts" setup>
import LanguageSelector from '@comp/LanguageSelector.vue';
import LoadingTextarea from '@comp/LoadingTextarea.vue';
import { useTranslation } from '@use/useTranslation';
import LayoutFull from '@lay/LayoutFull.vue';

const { pending, translation } = useTranslation();
</script>

<template>
    <LayoutFull>
        <LanguageSelector class="shrink-0" />
        <LoadingTextarea class="flex-1" :content="translation" :loading="pending"></LoadingTextarea>
    </LayoutFull>
</template>
