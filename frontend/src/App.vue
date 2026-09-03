<script lang="ts" setup>
import { useApplication } from '@use/useApplication';
import { onMounted, onUnmounted } from 'vue';
import { component } from '@/router';

const { start, stop } = useApplication();

onMounted(start);
onUnmounted(stop);
</script>

<template>
    <component :is="component" />
</template>
