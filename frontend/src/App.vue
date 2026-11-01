<script lang="ts" setup>
import { useApplication } from '@use/useApplication';
import { onMounted, onUnmounted } from 'vue';
import { RouterView } from 'vue-router';

const { start, stop } = useApplication();

onMounted(start);
onUnmounted(stop);
</script>

<template>
    <RouterView />
</template>
