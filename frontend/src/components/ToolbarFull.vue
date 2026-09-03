<script lang="ts" setup>
import Button, { type ButtonProps } from '@comp/Button.vue';
import ButtonGroup from '@comp/ButtonGroup.vue';
import ButtonGroupSlide from '@comp/ButtonGroupSlide.vue';
import ToolbarBase from '@comp/ToolbarBase.vue';
import { Check, ChevronLeft, ChevronRight, Clipboard } from '@lucide/vue';
import { useHistory } from '@use/useHistory';
import { useTranslation } from '@use/useTranslation';
import { computed, onUnmounted, ref } from 'vue';

const props = defineProps<{
    class?: string;
}>();

const { copyTranslation, translation } = useTranslation();
const { previousTranslation, nextTranslation, hasPrevious, hasNext } = useHistory();

/**
 * How long the check mark stands after copying, and how long it takes to fade out afterwards.
 */
const checkDuration = 1500;
const checkFadeDuration = 500;

/**
 * Whether the check mark is shown at all, and whether it is on its way out. Two flags and not one
 * state, because the icon stays in place while it fades and only then turns back into the clipboard.
 */
const copied = ref(false);
const fading = ref(false);

let copiedTimeout: ReturnType<typeof setTimeout>;

/**
 * The classes that animate the check mark, or nothing at all while the clipboard icon is shown.
 */
const copyIconClass = computed(() => {
    if (!copied.value) return undefined;

    return fading.value ? 'animate-check-out text-(--text-success)' : 'animate-check text-(--text-success)';
});

/**
 * Copies the translation and confirms it on the button. Restarts the confirmation when it is pressed
 * again while the check mark is still standing.
 */
async function copy(): Promise<void> {
    await copyTranslation();

    clearTimeout(copiedTimeout);
    copied.value = true;
    fading.value = false;

    copiedTimeout = setTimeout(() => {
        fading.value = true;

        copiedTimeout = setTimeout(() => {
            copied.value = false;
            fading.value = false;
        }, checkFadeDuration);
    }, checkDuration);
}

onUnmounted(() => {
    clearTimeout(copiedTimeout);
});

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
                :icon="copied ? Check : Clipboard"
                :title="copied ? 'Copied to Clipboard' : 'Copy to Clipboard'"
                style="--wails-draggable: no-drag"
            />
        </template>
    </ToolbarBase>
</template>
