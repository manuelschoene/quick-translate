<script lang="ts" setup>
import { Pencil } from '@lucide/vue';

defineProps<{
    action: () => void;
    label: string;
}>();
</script>

<template>
    <button
        class="group flex w-full items-center justify-between overflow-hidden rounded-lg border border-white/10 bg-linear-to-br from-gray-600 to-gray-700 p-2 hover:cursor-pointer hover:from-gray-500 hover:to-gray-600 focus-visible:ring-2 focus-visible:ring-gray-400 focus-visible:outline-none focus-visible:ring-inset"
        :title="label"
        @click="action"
        type="button"
    >
        <span class="truncate text-sm text-gray-50/80 transition-colors group-hover:text-gray-50">
            <slot />
        </span>
        <Pencil class="size-4 stroke-2 transition-transform group-hover:scale-125" />
    </button>
</template>
