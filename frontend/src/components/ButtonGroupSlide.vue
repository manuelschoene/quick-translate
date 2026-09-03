<script lang="ts" setup>
import Button from '@comp/Button.vue';
import { useProviders } from '@use/useProviders';
import { computed } from 'vue';

const { provider, alternatives, changeProvider } = useProviders();

/**
 * Whether there is anything to switch to. A single configured provider leaves the button as a label
 * that does not react, instead of sliding open on an empty list.
 */
const hasAlternatives = computed(() => alternatives.value.length > 0);
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
            :icon="provider.icon"
            :tabindex="hasAlternatives ? undefined : -1"
            :title="`Provider: ${provider.label}`"
            inline
        />

        <div
            class="grid grid-cols-[0fr] transition-[grid-template-columns] duration-300 group-hover/container:grid-cols-[1fr]"
            v-if="hasAlternatives"
        >
            <div class="flex min-w-0 divide-x divide-white/10 overflow-hidden">
                <Button
                    class="transition-colors hover:bg-linear-to-br hover:from-gray-500 hover:to-gray-600"
                    :action="() => changeProvider(alternative.slug)"
                    :icon="alternative.icon"
                    :key="alternative.slug"
                    :title="`Switch to ${alternative.label}`"
                    inline
                    v-for="alternative in alternatives"
                />
            </div>
        </div>
    </div>
</template>
