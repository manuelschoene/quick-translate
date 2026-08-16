import { copyTranslation } from '@data/actions';
import { requestState, translationState } from '@data/state';
import { view } from '@data/view';

const translation = view(translationState);
const { pending } = view(requestState);

/**
 * Gives the components the text that was translated, what came back for it and the way to take the
 * result over. Both are empty until the first translation was made.
 */
export function useTranslation() {
    return { ...translation, pending, copyTranslation };
}
