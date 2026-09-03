<script lang="ts" setup>
import { cn } from '@lib/cn';
import { LoaderCircle } from '@lucide/vue';

const props = defineProps<{
    content: string;
    loading: boolean;
    placeholder: string;
    class?: string;
}>();
</script>

<template>
    <div :class="cn('relative w-full p-2', props.class)">
        <textarea
            class="h-full w-full resize-none overflow-auto rounded-lg border border-white/10 bg-linear-to-br from-gray-700 to-gray-800 p-2 text-sm transition-colors placeholder:text-(--text-muted) focus-visible:outline-none"
            :class="{ 'text-(--text-muted)': props.loading }"
            :placeholder="props.loading ? '' : props.placeholder"
            :value="props.content"
            readonly
            tabindex="-1"
        />

        <div class="absolute inset-0 flex items-center justify-center" v-if="props.loading">
            <LoaderCircle class="size-6 animate-spin text-(--accent-default)" />
        </div>
    </div>
</template>
