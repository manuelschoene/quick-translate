import ErrorView from '@views/ErrorView.vue';
import SelectionSourceView from '@views/SelectionSourceView.vue';
import SelectionTargetView from '@views/SelectionTargetView.vue';
import TranslationView from '@views/TranslationView.vue';
import { ref, type Component } from 'vue';

/**
 * Represents the different views that the application can display. Each view name corresponds to a specific component that will be rendered when the application routes to that view.
 */
export type ViewName = 'translation' | 'error' | 'sourceSelection' | 'targetSelection';

const routes: Record<ViewName, Component> = {
    translation: TranslationView,
    error: ErrorView,
    sourceSelection: SelectionSourceView,
    targetSelection: SelectionTargetView,
};

/**
 * Holds the current component to be displayed as a view in the window.
 */
export const component = ref<Component | null>(TranslationView);

/**
 * Routes the application to the specified view. The component will be mounted newly and not be cached.
 * This ensures that each time a view is navigated to, it is freshly instantiated.
 * @param name The view name to route to.
 */
export function route(name: ViewName): void {
    component.value = routes[name];
}
