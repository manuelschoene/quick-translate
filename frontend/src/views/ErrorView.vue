<script lang="ts" setup>
import { CircleX } from '@lucide/vue';
import LayoutLabelled from '@lay/LayoutLabelled.vue';
import { useError } from '@use/useError';
import { onKey } from '@utils/keyboard';

const { message, showTranslation } = useError();

onKey('Escape', showTranslation);
</script>

<template>
    <LayoutLabelled :back-action="showTranslation" label="Error">
        <div class="h-full w-full overflow-y-auto bg-linear-to-br from-gray-800 to-gray-900">
            <div class="flex min-h-full flex-col items-center justify-center gap-4 px-4 py-4">
                <CircleX class="size-12 shrink-0 text-(--text-error)" />
                <span class="text-center">{{ message }}</span>
            </div>
        </div>
    </LayoutLabelled>
</template>
