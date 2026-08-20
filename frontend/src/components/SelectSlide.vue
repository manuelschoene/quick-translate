<script lang="ts" setup>
import ButtonIcon from '@comp/ButtonIcon.vue';
import { alternativeProviders, currentProvider } from '@utils/provider';
import { useProviders } from '@use/useProviders';

const current = currentProvider();
const alternative = alternativeProviders();
const { changeProvider } = useProviders();
</script>

<template>
    <div
        class="group/container flex items-center divide-x divide-white/10 overflow-hidden rounded-lg border-white/10 bg-linear-to-br from-gray-600 to-gray-700 hover:border"
    >
        <ButtonIcon
            class="shrink-0 transition-colors group-hover/container:rounded-none group-hover/container:bg-none hover:bg-linear-to-br hover:from-gray-500 hover:to-gray-600"
            :action="() => {}"
            :class="
                alternative.length > 0
                    ? 'group-hover/container:border-y-0 group-hover/container:border-r group-hover/container:border-l-0'
                    : 'group-hover/container:border-none'
            "
            :icon="current.icon"
            :label="current.label"
        />

        <div
            class="grid grid-cols-[0fr] transition-[grid-template-columns] duration-300 group-hover/container:grid-cols-[1fr]"
        >
            <div class="flex min-w-0 divide-x divide-white/10 overflow-hidden">
                <button
                    class="group/button flex items-center justify-center p-2 transition-colors hover:cursor-pointer hover:bg-linear-to-br hover:from-gray-500 hover:to-gray-600 focus-visible:ring-2 focus-visible:ring-gray-400 focus-visible:outline-none focus-visible:ring-inset"
                    :key="provider.slug"
                    :title="provider.label"
                    @click="changeProvider(provider.slug)"
                    type="button"
                    v-for="provider in alternative"
                >
                    <component
                        class="size-4 stroke-2 transition-transform group-hover/button:scale-125"
                        :is="provider.icon"
                    />
                </button>
            </div>
        </div>
    </div>
</template>
