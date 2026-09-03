import ErrorView from '@views/ErrorView.vue';
import SelectionSourceView from '@views/SelectionSourceView.vue';
import SelectionTargetView from '@views/SelectionTargetView.vue';
import TranslationView from '@views/TranslationView.vue';
import { shallowRef, type Component } from 'vue';

export type ViewName = 'translation' | 'error' | 'sourceSelection' | 'targetSelection';

/**
 * The only place that decides which component a view name stands for. Typed against `ViewName`, so a
 * name without a component does not compile.
 */
const routes: Record<ViewName, Component> = {
    translation: TranslationView,
    error: ErrorView,
    sourceSelection: SelectionSourceView,
    targetSelection: SelectionTargetView,
};

/**
 * The view that is on screen. It starts on the translation, which is what the window is opened for,
 * and there is no state without one. Shallow, because a component definition is held as it is and
 * not turned into a reactive proxy.
 */
export const component = shallowRef<Component>(TranslationView);

/**
 * Takes the user to the given view. The component is mounted anew and not kept around, so a view
 * always starts fresh.
 */
export function route(name: ViewName): void {
    component.value = routes[name];
}
