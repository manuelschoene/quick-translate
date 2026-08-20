<script lang="ts" setup>
import LayoutLabelled from '@lay/LayoutLabelled.vue';
import { useError } from '@use/useError';
import { CircleX } from '@lucide/vue';

const { message, showTranslation } = useError();
</script>

<template>
    <LayoutLabelled :back-action="showTranslation" label="Error">
        <div class="flex h-full w-full items-center justify-center bg-linear-to-br from-gray-800 to-gray-900 px-4">
            <div class="flex flex-col items-center gap-4">
                <CircleX class="size-16 text-error-500" />
                <span class="text-center">{{ message }}</span>
            </div>
        </div>
    </LayoutLabelled>
</template>
