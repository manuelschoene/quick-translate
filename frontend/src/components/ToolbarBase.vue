<script lang="ts" setup>
import Panel from '@comp/Panel.vue';
import ButtonIcon from '@comp/ButtonIcon.vue';
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
            <ButtonIcon :action="hide" :icon="X" label="Close Window" style="--wails-draggable: no-drag" />
        </div>
    </Panel>
</template>
