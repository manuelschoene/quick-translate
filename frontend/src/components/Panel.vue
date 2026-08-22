<script lang="ts" setup>
import { cn } from '@utils/cn';

const props = defineProps<{
    class?: string;
}>();
</script>

<template>
    <div :class="cn('border-white/10 bg-linear-to-br from-gray-700 to-gray-800 p-2', props.class)">
        <slot />
    </div>
</template>
