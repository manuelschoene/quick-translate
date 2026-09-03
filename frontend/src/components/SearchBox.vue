<script lang="ts" setup>
import { onMounted, ref, useTemplateRef, watch } from 'vue';
import { X } from '@lucide/vue';
import Button from '@comp/Button.vue';

const props = defineProps<{
    placeholder?: string;
    modelValue: string;
    label?: string;
}>();

const emit = defineEmits<{
    'update:modelValue': [value: string];
    submit: [];
}>();

const local = ref(props.modelValue);
let timeout: ReturnType<typeof setTimeout>;

watch(local, (newValue) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => {
        emit('update:modelValue', newValue);
    }, 100);
});

const submit = (): void => {
    clearTimeout(timeout);
    emit('update:modelValue', local.value);
    emit('submit');
};

const clear = (): void => {
    local.value = '';
};

const input = useTemplateRef<HTMLInputElement>('input');
onMounted(() => input.value?.focus());
</script>

<template>
    <div
        class="group/search relative flex h-(--control-height) items-center overflow-hidden rounded-lg border border-white/10 bg-linear-to-br from-gray-600 to-gray-700 transition-colors"
    >
        <input
            class="h-full w-full bg-transparent p-2 text-sm outline-none placeholder:text-(--text-muted)"
            id="search-input"
            :class="{ 'pr-8': local }"
            :placeholder="placeholder"
            :title="label ?? 'Search'"
            @keydown.enter="submit"
            ref="input"
            type="text"
            v-model="local"
        />

        <Button
            class="absolute right-0 h-full py-0 opacity-50 transition-colors hover:bg-linear-to-br hover:from-gray-500 hover:to-gray-600 hover:opacity-100"
            :action="clear"
            :icon="X"
            inline
            title="Clear Search"
            v-if="local"
        />

        <div
            class="pointer-events-none absolute inset-0 rounded-lg ring-(--accent-default) ring-inset group-has-[input:focus]/search:ring-2"
        />
    </div>
</template>
