<script lang="ts" setup>
import Button from '@comp/Button.vue';
import Panel from '@comp/Panel.vue';
import { cn } from '@lib/cn';
import { X } from '@lucide/vue';
import { useApplication } from '@use/useApplication';

const props = defineProps<{
    class?: string;
}>();

const { hide } = useApplication();
</script>

<template>
    <Panel :class="cn('flex items-center justify-between border-t', props.class)" style="--wails-draggable: drag">
        <div class="flex">
            <slot name="left" />
        </div>

        <div class="flex gap-2">
            <slot name="right" />
            <Button :action="hide" :icon="X" style="--wails-draggable: no-drag" title="Close Window" />
        </div>
    </Panel>
</template>
