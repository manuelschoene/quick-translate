<script lang="ts">
/**
 * Everything a button of the application is built from. A button carries a text, an icon or both,
 * and `inline` drops the frame it draws around itself so that it can be placed inside a container
 * that already brings the height, the border and the background along. `class`, `classText` and
 * `classIcon` are merged onto the button, onto its text and onto its icon by the call site, which is
 * how the few buttons that need to stretch, truncate, fade or animate are told apart from the rest.
 *
 * Exported here and not from a file of its own, because a container that lays out buttons describes
 * them with the same type the button itself is declared with.
 */
export interface ButtonProps {
    action: () => void;
    title: string;
    text?: string;
    icon?: Component;
    disabled?: boolean;
    inline?: boolean;
    class?: string;
    classText?: string;
    classIcon?: string;
}
</script>

<script lang="ts" setup>
import { cn } from '@lib/cn';
import { computed, type Component } from 'vue';

const props = defineProps<ButtonProps>();

const buttonClass = computed(() =>
    cn(
        'group flex items-center justify-center px-2 py-2 hover:cursor-pointer focus-visible:ring-2 focus-visible:ring-(--accent-default) focus-visible:outline-none focus-visible:ring-inset disabled:cursor-not-allowed disabled:hover:cursor-not-allowed',
        !props.inline
            && 'h-(--control-height) overflow-hidden rounded-lg border border-white/10 bg-linear-to-br from-gray-600 to-gray-700 py-0 transition-colors hover:from-gray-500 hover:to-gray-600 disabled:hover:from-gray-600 disabled:hover:to-gray-700',
        props.text && props.icon && 'justify-between gap-2',
        props.icon && !props.text && 'shrink-0',
        props.class,
    ),
);

const textClass = computed(() =>
    cn(
        'text-sm text-(--text-secondary) transition-colors group-hover:text-(--text-primary) group-hover:[text-shadow:0.25px_0_currentColor,-0.25px_0_currentColor] group-disabled:text-(--text-muted) group-disabled:group-hover:[text-shadow:none]',
        props.classText,
    ),
);

const iconClass = computed(() =>
    cn(
        'size-4 shrink-0 stroke-2 transition-transform group-hover:scale-125 group-disabled:scale-100 group-disabled:text-(--text-muted)',
        props.classIcon,
    ),
);
</script>

<template>
    <button :class="buttonClass" :disabled="props.disabled" :title="props.title" @click="props.action" type="button">
        <span :class="textClass" v-if="props.text">{{ props.text }}</span>

        <component :class="iconClass" :is="props.icon" v-if="props.icon" />
    </button>
</template>
