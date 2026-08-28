<script lang="ts" setup>
import { LoaderCircle } from '@lucide/vue';
import { computed } from 'vue';

const props = defineProps<{
    content?: string;
    loading?: boolean;
    placeholder?: string;
}>();

const value = computed(() => props.content ?? '');
</script>

<template>
    <div class="relative w-full p-2">
        <textarea
            class="h-full w-full resize-none overflow-auto rounded-lg border border-white/10 bg-linear-to-br from-gray-700 to-gray-800 p-2 text-sm transition-colors placeholder:text-(--text-muted) focus-visible:outline-none"
            :class="{ 'text-(--text-muted)': loading }"
            :placeholder="loading ? '' : placeholder"
            :value="value"
            readonly
            tabindex="-1"
        />

        <div class="absolute inset-0 flex items-center justify-center" v-if="loading">
            <LoaderCircle class="size-6 animate-spin text-(--accent-default)" />
        </div>
    </div>
</template>
