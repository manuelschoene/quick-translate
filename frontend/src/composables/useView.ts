import { viewState } from '@data/state';
import { showError, showTranslation } from '@data/views';
import { registry } from '@views/registry';
import { computed } from 'vue';

const entry = computed(() => registry[viewState.current]);
const view = computed(() => entry.value.view);
const layout = computed(() => entry.value.layout);

/**
 * Gives what is to be rendered and the ways to change it. The view and its layout are handed out
 * separately, so the application renders the one inside the other.
 *
 * Both ways to switch are open to the components and to the data layer, which uses them when the
 * backend reports a translation or an error on its own.
 */
export function useView() {
    return { view, layout, showTranslation, showError };
}
