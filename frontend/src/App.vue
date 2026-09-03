<script lang="ts" setup>
import { component } from '@/router';
import { useApplication } from '@use/useApplication';
import { onMounted, onUnmounted } from 'vue';

const { start, stop } = useApplication();

onMounted(start);
onUnmounted(stop);
</script>

<template>
    <component :is="component" />
</template>
