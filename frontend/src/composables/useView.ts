import { showError, showTranslation } from '@data/navigation';
import { viewState } from '@data/state';
import { router } from '@views/router';
import { computed } from 'vue';

const view = computed(() => router[viewState.current]);

/**
 * Gives the view that is to be rendered and the ways to change it. The view is handed out on its
 * own, because it brings its layout along by importing it.
 *
 * Both ways to switch are open to the components and to the data layer, which uses them when the
 * backend reports a translation or an error on its own.
 */
export function useView() {
    return { view, showTranslation, showError };
}
