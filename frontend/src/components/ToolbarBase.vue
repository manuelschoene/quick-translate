<script lang="ts" setup>
import Button from '@comp/Button.vue';
import Panel from '@comp/Panel.vue';
import { cn } from '@lib/cn';
import { feedback } from '@lib/feedback';
import { X } from '@lucide/vue';
import { useApplication } from '@use/useApplication';
import { computed } from 'vue';

const props = defineProps<{
    class?: string;
}>();

const { hide } = useApplication();

const { outcome: hideOutcome, iconClass: hideIconClass, show: showHideOutcome } = feedback();

/**
 * What the close button says it does, or that it could not do it. The icon stays the cross it always
 * was, the answer is carried by its colour and the way it comes in.
 */
const hideTitle = computed(() => (hideOutcome.value === 'failure' ? 'Could Not Close Window' : 'Close Window'));

/**
 * Closes the window and answers on the button when that did not work.
 *
 * Only a failure is answered: a window that did close takes the answer with it, and one that is
 * opened again by the shortcut a moment later would greet the user with a confirmation of something
 * they no longer remember asking for.
 */
async function dismiss(): Promise<void> {
    const hidden = await hide();

    if (!hidden) {
        showHideOutcome(false);
    }
}
</script>

<template>
    <Panel :class="cn('flex items-center justify-between border-t', props.class)" style="--wails-draggable: drag">
        <div class="flex">
            <slot name="left" />
        </div>

        <div class="flex gap-2">
            <slot name="right" />
            <Button
                :action="dismiss"
                :class-icon="hideIconClass"
                :icon="X"
                :title="hideTitle"
                style="--wails-draggable: no-drag"
            />
        </div>
    </Panel>
</template>
