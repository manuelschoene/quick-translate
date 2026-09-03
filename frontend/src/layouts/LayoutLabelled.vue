<script lang="ts" setup>
import ToolbarLabelled from '@comp/ToolbarLabelled.vue';
import LayoutBase from '@lay/LayoutBase.vue';

const props = defineProps<{
    label: string;
    backAction: () => void;
}>();
</script>

<template>
    <LayoutBase>
        <div class="flex min-h-0 flex-1 flex-col">
            <slot />
        </div>

        <ToolbarLabelled class="shrink-0" :back-action="props.backAction" :label="props.label" />
    </LayoutBase>
</template>
