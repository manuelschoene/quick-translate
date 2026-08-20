<script lang="ts" setup>
import ButtonIcon from '@comp/ButtonIcon.vue';
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
                <ButtonIcon :action="backAction" :icon="ArrowLeft" label="Back" style="--wails-draggable: no-drag" />
                <span class="text-sm font-semibold">{{ label }}</span>
            </div>
        </template>
    </ToolbarBase>
</template>
