<script lang="ts" setup>
import Button from '@comp/Button.vue';
import { X } from '@lucide/vue';
import { onMounted, ref, useTemplateRef, watch } from 'vue';

const props = defineProps<{
    placeholder?: string;
    modelValue: string;
    title?: string;
}>();

const emit = defineEmits<{
    'update:modelValue': [value: string];
    submit: [];
}>();

/**
 * What is currently in the box, which runs ahead of the value that was handed out.
 */
const local = ref('');

let timeout: ReturnType<typeof setTimeout>;

/**
 * Follows the value that is handed in, which is also what fills the box the first time.
 *
 * Registered before the watcher below, so the first run happens while there is nothing listening yet
 * and a value that arrives with the component is not sent straight back.
 */
watch(
    () => props.modelValue,
    (value) => {
        local.value = value;
    },
    { immediate: true },
);

/**
 * Hands what was typed back out, but not on every keystroke: the list that is searched is rebuilt
 * for every value, so the box waits until the typing pauses.
 */
watch(local, (newValue) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => {
        emit('update:modelValue', newValue);
    }, 100);
});

/**
 * Hands the value out right away and asks for it to be acted on, which is what pressing enter means.
 * Cancels the pending debounce, so the value is not sent a second time.
 */
function submit(): void {
    clearTimeout(timeout);
    emit('update:modelValue', local.value);
    emit('submit');
}

/**
 * Empties the box. The value leaves debounced like any other change.
 */
function clear(): void {
    local.value = '';
}

const input = useTemplateRef<HTMLInputElement>('input');

/**
 * Puts the caret back into the box. Handed out so that a view which took the focus away can give it
 * back without reaching into the input itself.
 */
function focus(): void {
    input.value?.focus();
}

// The box is the reason its view was opened, so the user can type without reaching for it first.
onMounted(focus);

defineExpose({ focus });
</script>

<template>
    <div
        class="group/search relative flex h-(--control-height) items-center overflow-hidden rounded-lg border border-white/10 bg-linear-to-br from-gray-600 to-gray-700 transition-colors"
    >
        <input
            class="h-full w-full bg-transparent p-2 text-sm outline-none placeholder:text-(--text-muted)"
            :class="{ 'pr-8': local }"
            :placeholder="props.placeholder"
            :title="props.title ?? 'Search'"
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
