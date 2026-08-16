<script lang="ts" setup>
import { useApplication } from '@use/useApplication';
import { useView } from '@use/useView';
import { onMounted, onUnmounted } from 'vue';

const { start, stop } = useApplication();
const { layout, view } = useView();

onMounted(start);
onUnmounted(stop);
</script>

<template>
    <component :is="layout">
        <component :is="view" />
    </component>
</template>
