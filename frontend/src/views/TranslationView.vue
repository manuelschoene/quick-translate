<script lang="ts" setup>
import Button from '@comp/Button.vue';
import LoadingTextarea from '@comp/LoadingTextarea.vue';
import Panel from '@comp/Panel.vue';
import LayoutFull from '@lay/LayoutFull.vue';
import { ArrowRightLeft, Pencil } from '@lucide/vue';
import { useApplication } from '@use/useApplication';
import { useLanguages } from '@use/useLanguages';
import { useTranslation } from '@use/useTranslation';
import { onKey } from '@lib/keyboard';
import { route } from '@/router';

const { pending, translation } = useTranslation();
const { sourceLabel, targetLabel, switchLanguages } = useLanguages();

const { hide } = useApplication();

onKey('Escape', hide);
</script>

<template>
    <LayoutFull>
        <Panel class="flex shrink-0 justify-between gap-2 border-b" style="--wails-draggable: drag">
            <Button
                class="w-full"
                :action="() => route('sourceSelection')"
                :icon="Pencil"
                :text="sourceLabel"
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
                :action="() => route('targetSelection')"
                :icon="Pencil"
                :text="targetLabel"
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
