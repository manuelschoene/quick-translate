import { nextTranslation, previousTranslation } from '@data/actions';
import { historyState, requestState } from '@data/state';
import { view } from '@data/view';

const history = view(historyState);
const { pending } = view(requestState);

/**
 * Gives the components the ways to step through the translations that were made before and whether
 * there is anything to step to. Both flags stay false while the history is turned off.
 *
 * A stored translation is shown as it was made and is not translated again. It brings its provider
 * and its languages with it, which the other composables pick up on their own.
 */
export function useHistory() {
    return { ...history, pending, previousTranslation, nextTranslation };
}
