<script lang="ts" setup>
import ButtonIcon from '@comp/ButtonIcon.vue';
import ButtonIconGroup from '@comp/ButtonIconGroup.vue';
import { ChevronLeft, ChevronRight, Clipboard } from '@lucide/vue';
import ToolbarBase from '@comp/ToolbarBase.vue';
import { useTranslation } from '@use/useTranslation';
import { useHistory } from '@use/useHistory';
import SelectSlide from '@comp/SelectSlide.vue';

const { copyTranslation } = useTranslation();
const { previousTranslation, nextTranslation, hasPrevious, hasNext } = useHistory();

const navigationButtons = [
    {
        action: previousTranslation,
        icon: ChevronLeft,
        label: 'Last Translation',
        disabled: !hasPrevious,
    },
    {
        action: nextTranslation,
        icon: ChevronRight,
        label: 'Next Translation',
        disabled: !hasNext,
    },
];
</script>

<template>
    <ToolbarBase>
        <template #left>
            <SelectSlide style="--wails-draggable: no-drag" />
        </template>

        <template #right>
            <ButtonIconGroup :buttons="navigationButtons" style="--wails-draggable: no-drag" />
            <ButtonIcon
                :action="copyTranslation"
                :icon="Clipboard"
                label="Copy to Clipboard"
                style="--wails-draggable: no-drag"
            />
        </template>
    </ToolbarBase>
</template>
