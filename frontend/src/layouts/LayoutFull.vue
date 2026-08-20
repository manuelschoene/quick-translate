<script lang="ts" setup>
import ToolbarFull from '@comp/ToolbarFull.vue';
import LayoutBase from '@lay/LayoutBase.vue';
</script>

<template>
    <LayoutBase>
        <div class="flex flex-1 flex-col">
            <slot />
        </div>

        <ToolbarFull class="shrink-0" />
    </LayoutBase>
</template>
