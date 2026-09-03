<script lang="ts" setup>
import Button, { type ButtonProps } from '@comp/Button.vue';
import ButtonGroup from '@comp/ButtonGroup.vue';
import ButtonGroupSlide from '@comp/ButtonGroupSlide.vue';
import ToolbarBase from '@comp/ToolbarBase.vue';
import { feedback } from '@lib/feedback';
import { Check, ChevronLeft, ChevronRight, Clipboard, X } from '@lucide/vue';
import { useHistory } from '@use/useHistory';
import { useTranslation } from '@use/useTranslation';
import { computed } from 'vue';

const props = defineProps<{
    class?: string;
}>();

const { copyTranslation, translation } = useTranslation();
const { previousTranslation, nextTranslation, hasPrevious, hasNext } = useHistory();

const { outcome: copyOutcome, iconClass: copyIconClass, show: showCopyOutcome } = feedback();

/**
 * The icon of the copy button, which answers the last press before it turns back into the clipboard.
 */
const copyIcon = computed(() => {
    switch (copyOutcome.value) {
        case 'success':
            return Check;
        case 'failure':
            return X;
        default:
            return Clipboard;
    }
});

/**
 * What the copy button says it does, or what it made of the last press.
 */
const copyTitle = computed(() => {
    switch (copyOutcome.value) {
        case 'success':
            return 'Copied to Clipboard';
        case 'failure':
            return 'Could Not Copy to Clipboard';
        default:
            return 'Copy to Clipboard';
    }
});

/**
 * Copies the translation and answers on the button, because a failure is reported quietly and would
 * otherwise leave the press looking like it worked.
 */
async function copy(): Promise<void> {
    showCopyOutcome(await copyTranslation());
}

/**
 * The two buttons for stepping through the history, each turned off when there is nothing to step to.
 */
const navigationButtons = computed<ButtonProps[]>(() => [
    {
        action: previousTranslation,
        icon: ChevronLeft,
        title: 'Last Translation',
        disabled: !hasPrevious.value,
    },
    {
        action: nextTranslation,
        icon: ChevronRight,
        title: 'Next Translation',
        disabled: !hasNext.value,
    },
]);
</script>

<template>
    <ToolbarBase :class="props.class">
        <template #left>
            <ButtonGroupSlide style="--wails-draggable: no-drag" />
        </template>

        <template #right>
            <ButtonGroup :buttons="navigationButtons" style="--wails-draggable: no-drag" />
            <Button
                :action="copy"
                :class-icon="copyIconClass"
                :disabled="!translation"
                :icon="copyIcon"
                :title="copyTitle"
                style="--wails-draggable: no-drag"
            />
        </template>
    </ToolbarBase>
</template>
