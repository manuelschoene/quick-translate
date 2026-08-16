import { readonly, toRefs, type DeepReadonly, type Ref } from 'vue';

type View<T extends object> = { readonly [K in keyof T]: Readonly<Ref<DeepReadonly<T[K]>>> };

/**
 * Turns a state group into the refs the composables hand out. They keep their reactivity when a
 * component destructures them and refuse to be written to, so the state changes through the actions
 * alone.
 *
 * The cast is needed because `toRefs` types its refs as writable even over a read only state, which
 * would leave a write to be caught at runtime instead of while typing it.
 */
export function view<T extends object>(state: T): View<T> {
    return toRefs(readonly(state)) as View<T>;
}
