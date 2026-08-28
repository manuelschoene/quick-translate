<script lang="ts" setup>
import { Check, ChevronLeft, ChevronRight, Clipboard } from '@lucide/vue';
import ToolbarBase from '@comp/ToolbarBase.vue';
import { useTranslation } from '@use/useTranslation';
import { useHistory } from '@use/useHistory';
import ButtonGroupSlide from '@comp/ButtonGroupSlide.vue';
import Button from './Button.vue';
import ButtonGroup from './ButtonGroup.vue';
import type { ButtonProps } from '@utils/types';
import { computed, onUnmounted, ref } from 'vue';

const { copyTranslation, translation } = useTranslation();
const { previousTranslation, nextTranslation, hasPrevious, hasNext } = useHistory();

const checkDuration = 1500;
const checkFadeDuration = 500;

const copied = ref(false);
const fading = ref(false);
let copiedTimeout: ReturnType<typeof setTimeout>;

const copyIconClass = computed(() => {
    if (!copied.value) return undefined;

    return fading.value ? 'animate-check-out text-(--text-success)' : 'animate-check text-(--text-success)';
});

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

onUnmounted(() => clearTimeout(copiedTimeout));

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
    <ToolbarBase>
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
