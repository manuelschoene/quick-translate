<script lang="ts" setup>
import Panel from '@comp/Panel.vue';
import Button from '@comp/Button.vue';
import { X } from '@lucide/vue';
import { useApplication } from '@use/useApplication';

const { hide } = useApplication();
</script>

<template>
    <Panel class="flex items-center justify-between border-t" style="--wails-draggable: drag">
        <div class="flex">
            <slot name="left" />
        </div>

        <div class="flex gap-2">
            <slot name="right" />
            <Button :action="hide" :icon="X" style="--wails-draggable: no-drag" title="Close Window" />
        </div>
    </Panel>
</template>
