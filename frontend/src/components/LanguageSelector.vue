<script lang="ts" setup>
import { ArrowRightLeft } from '@lucide/vue';
import Panel from '@comp/Panel.vue';
import ButtonSelect from '@comp/ButtonSelect.vue';
import ButtonIcon from '@comp/ButtonIcon.vue';
import { useLanguages } from '@use/useLanguages.ts';
import { displaySourceLanguage, displayTargetLanguage } from '@utils/language.ts';

const { switchLanguages } = useLanguages();
const source = displaySourceLanguage();
const target = displayTargetLanguage();
</script>

<template>
    <Panel class="flex justify-between gap-2 border-b">
        <ButtonSelect :action="() => {}" label="Select Source Language">{{ source }}</ButtonSelect>

        <ButtonIcon :action="switchLanguages" :icon="ArrowRightLeft" label="Switch Languages" />

        <ButtonSelect :action="() => {}" label="Select Target Language">{{ target }}</ButtonSelect>
    </Panel>
</template>
