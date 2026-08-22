<script lang="ts" setup>
import { ref, watch } from 'vue';
import { X } from '@lucide/vue';

const props = defineProps<{
    placeholder?: string;
    modelValue: string;
    label?: string;
}>();

const emit = defineEmits<{
    'update:modelValue': [value: string];
}>();

const local = ref(props.modelValue);
let timeout: ReturnType<typeof setTimeout>;

watch(local, (newValue) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => {
        emit('update:modelValue', newValue);
    }, 100);
});

const clear = () => (local.value = '');
</script>

<template>
    <div
        class="relative flex items-center overflow-hidden rounded-lg border border-white/10 bg-linear-to-br from-gray-600 to-gray-700 transition-colors has-[input:focus]:ring-2 has-[input:focus]:ring-gray-400 has-[input:focus]:ring-inset"
    >
        <input
            class="h-full w-full bg-transparent p-2 text-sm outline-none placeholder:text-gray-400"
            id="search-input"
            :class="{ 'pr-8': local }"
            :placeholder="placeholder"
            :title="label ?? 'Search'"
            type="text"
            v-model="local"
            autofocus
        />

        <button
            class="group absolute right-0 flex h-full items-center justify-center p-2 transition-colors hover:cursor-pointer hover:bg-linear-to-br hover:from-gray-500 hover:to-gray-600 focus-visible:ring-2 focus-visible:ring-gray-400 focus-visible:outline-none focus-visible:ring-inset"
            @click="clear"
            title="Clear Search"
            type="button"
            v-if="local"
        >
            <X class="size-4 stroke-2 opacity-50 transition-transform group-hover:scale-125 group-hover:opacity-100" />
        </button>
    </div>
</template>
