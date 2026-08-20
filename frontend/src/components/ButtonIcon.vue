<script lang="ts" setup>
import { cn } from '@utils/cn';
import { type Component } from 'vue';

export interface ButtonIconProps {
    action: () => void;
    icon: Component;
    label: string;
    disabled?: boolean;
    class?: string;
}

const props = defineProps<ButtonIconProps>();
</script>

<template>
    <button
        :class="
            cn(
                'group flex items-center justify-center rounded-lg border border-white/10 bg-linear-to-br from-gray-600 to-gray-700 p-2 transition-colors hover:cursor-pointer hover:from-gray-500 hover:to-gray-600 focus-visible:ring-2 focus-visible:ring-gray-400 focus-visible:outline-none focus-visible:ring-inset disabled:cursor-not-allowed disabled:text-gray-50/30 disabled:hover:from-gray-600 disabled:hover:to-gray-700',
                props.class,
            )
        "
        :disabled="props.disabled"
        :title="props.label"
        @click="props.action"
        type="button"
    >
        <component
            class="size-4 stroke-2 transition-transform group-hover:scale-125 group-disabled:scale-100"
            :is="icon"
        />
    </button>
</template>
