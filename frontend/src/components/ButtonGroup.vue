<script lang="ts" setup>
import Button, { type ButtonProps } from '@comp/Button.vue';
import { cn } from '@lib/cn';

defineProps<{
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
            :key="button.title"
            :text="button.text"
            :title="button.title"
            inline
            v-for="button in buttons"
        />
    </div>
</template>
