<script lang="ts" setup>
import ToolbarLabelled from '@comp/ToolbarLabelled.vue';
import LayoutBase from '@lay/LayoutBase.vue';

defineProps<{
    label: string;
    backAction: () => void;
}>();
</script>

<template>
    <LayoutBase>
        <div class="flex flex-1 flex-col">
            <slot />
        </div>

        <ToolbarLabelled class="shrink-0" :back-action="backAction" :label="label" />
    </LayoutBase>
</template>
