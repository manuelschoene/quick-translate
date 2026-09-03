<script lang="ts" setup>
import Button from '@comp/Button.vue';
import ToolbarBase from '@comp/ToolbarBase.vue';
import { ArrowLeft } from '@lucide/vue';

const props = defineProps<{
    label: string;
    backAction: () => void;
    class?: string;
}>();
</script>

<template>
    <ToolbarBase :class="props.class">
        <template #left>
            <div class="flex items-center gap-4">
                <Button :action="props.backAction" :icon="ArrowLeft" style="--wails-draggable: no-drag" title="Back" />
                <span class="text-sm font-semibold">{{ props.label }}</span>
            </div>
        </template>
    </ToolbarBase>
</template>
