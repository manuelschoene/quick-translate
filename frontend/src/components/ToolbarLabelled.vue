<script lang="ts" setup>
import Button from '@comp/Button.vue';
import { ArrowLeft } from '@lucide/vue';
import ToolbarBase from '@comp/ToolbarBase.vue';

defineProps<{
    label: string;
    backAction: () => void;
}>();
</script>

<template>
    <ToolbarBase>
        <template #left>
            <div class="flex items-center gap-4">
                <Button :action="backAction" :icon="ArrowLeft" style="--wails-draggable: no-drag" title="Back" />
                <span class="text-sm font-semibold">{{ label }}</span>
            </div>
        </template>
    </ToolbarBase>
</template>
