<!--
    A placeholder that shows the state of the translation as it is. It stands in for the interface
    until that one is built and only proves that the data arrives.
-->
<script lang="ts" setup>
import { useTranslation } from '@use/useTranslation';

const { text, translation, pending } = useTranslation();
</script>

<template>
    <p v-if="pending">Translating …</p>
    <template v-else>
        <p>{{ text }}</p>
        <p>{{ translation }}</p>
    </template>
</template>
