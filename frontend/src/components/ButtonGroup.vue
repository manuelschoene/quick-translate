<script lang="ts" setup>
import Button, { type ButtonProps } from '@comp/Button.vue';
import { cn } from '@lib/cn';

/**
 * The buttons are keyed by their position, which holds as long as a call site writes the group out
 * as a fixed list. A group whose buttons are ever added, removed or reordered needs an identity of
 * its own on `ButtonProps` instead.
 */
const props = defineProps<{
    buttons: ButtonProps[];
}>();
</script>

<template>
    <div
        class="inline-flex h-(--control-height) divide-x divide-white/10 overflow-hidden rounded-lg border border-white/10 bg-linear-to-br from-gray-600 to-gray-700"
    >
        <Button
            :action="button.action"
            :class="
                cn(
                    'py-0 hover:bg-linear-to-br hover:from-gray-500 hover:to-gray-600 disabled:hover:bg-none',
                    button.class,
                )
            "
            :class-icon="button.classIcon"
            :class-text="button.classText"
            :disabled="button.disabled"
            :icon="button.icon"
            :key="index"
            :text="button.text"
            :title="button.title"
            inline
            v-for="(button, index) in props.buttons"
        />
    </div>
</template>
