<script lang="ts" setup>
import Button from '@comp/Button.vue';
import { alternativeProviders, currentProvider } from '@utils/provider';
import { useProviders } from '@use/useProviders';
import { computed } from 'vue';

const current = currentProvider();
const alternative = alternativeProviders();
const { changeProvider } = useProviders();

const hasAlternatives = computed(() => alternative.value.length > 0);
</script>

<template>
    <div
        class="group/container flex h-(--control-height) items-center divide-x divide-white/10 overflow-hidden rounded-lg border border-white/10 bg-linear-to-br from-gray-600 to-gray-700"
    >
        <Button
            :action="() => {}"
            :class="
                hasAlternatives
                    ? 'transition-colors hover:bg-linear-to-br hover:from-gray-500 hover:to-gray-600'
                    : 'hover:cursor-default'
            "
            :class-icon="hasAlternatives ? undefined : 'group-hover:scale-100'"
            :icon="current.icon"
            :tabindex="hasAlternatives ? undefined : -1"
            :title="`Provider: ${current.label}`"
            inline
        />

        <div
            class="grid grid-cols-[0fr] transition-[grid-template-columns] duration-300 group-hover/container:grid-cols-[1fr]"
            v-if="hasAlternatives"
        >
            <div class="flex min-w-0 divide-x divide-white/10 overflow-hidden">
                <Button
                    class="transition-colors hover:bg-linear-to-br hover:from-gray-500 hover:to-gray-600"
                    :action="() => changeProvider(provider.slug)"
                    :icon="provider.icon"
                    :key="provider.slug"
                    :title="`Switch to ${provider.label}`"
                    inline
                    v-for="provider in alternative"
                />
            </div>
        </div>
    </div>
</template>
