<script lang="ts" setup>
import { ButtonIconProps } from '@comp/ButtonIcon.vue';

defineProps<{
    buttons: ButtonIconProps[];
}>();
</script>

<template>
    <div
        class="inline-flex divide-x divide-white/10 overflow-hidden rounded-lg border border-white/10 bg-linear-to-br from-gray-600 to-gray-700"
    >
        <button
            class="group flex items-center justify-center p-2 transition-colors hover:cursor-pointer hover:bg-linear-to-br hover:from-gray-500 hover:to-gray-600 focus-visible:ring-2 focus-visible:ring-gray-400 focus-visible:outline-none focus-visible:ring-inset disabled:cursor-not-allowed disabled:text-gray-50/30 disabled:hover:bg-none"
            :disabled="buttonProps.disabled"
            :key="buttonProps.label"
            :title="buttonProps.label"
            @click="buttonProps.action"
            type="button"
            v-for="buttonProps in buttons"
        >
            <component
                class="size-4 stroke-2 transition-transform group-hover:scale-125 group-disabled:scale-100"
                :is="buttonProps.icon"
            />
        </button>
    </div>
</template>
