<script lang="ts" setup>
import Button from '@comp/Button.vue';
import LoadingTextarea from '@comp/LoadingTextarea.vue';
import Panel from '@comp/Panel.vue';
import LayoutFull from '@lay/LayoutFull.vue';
import { ArrowRightLeft, Pencil } from '@lucide/vue';
import { useApplication } from '@use/useApplication';
import { useLanguages } from '@use/useLanguages';
import { useTranslation } from '@use/useTranslation';
import { useView } from '@use/useView';
import { onKey } from '@utils/keyboard';
import { displaySourceLanguage, displayTargetLanguage } from '@utils/language';

const { pending, translation } = useTranslation();
const { switchLanguages } = useLanguages();
const { showSourceSelection, showTargetSelection } = useView();

const { hide } = useApplication();

const source = displaySourceLanguage();
const target = displayTargetLanguage();

onKey('Escape', hide);
</script>

<template>
    <LayoutFull>
        <Panel class="flex shrink-0 justify-between gap-2 border-b" style="--wails-draggable: drag">
            <Button
                class="w-full"
                :action="showSourceSelection"
                :icon="Pencil"
                :text="source"
                class-text="truncate"
                style="--wails-draggable: no-drag"
                title="Select Source Language"
            />

            <Button
                :action="switchLanguages"
                :icon="ArrowRightLeft"
                style="--wails-draggable: no-drag"
                title="Switch Languages"
            />

            <Button
                class="w-full"
                :action="showTargetSelection"
                :icon="Pencil"
                :text="target"
                class-text="truncate"
                style="--wails-draggable: no-drag"
                title="Select Target Language"
            />
        </Panel>

        <LoadingTextarea
            class="flex-1"
            :content="translation"
            :loading="pending"
            placeholder="Select a text anywhere and press your shortcut to translate it."
        />
    </LayoutFull>
</template>
