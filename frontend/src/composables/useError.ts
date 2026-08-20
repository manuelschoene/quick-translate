import { showTranslation } from '@data/navigation';
import { errorState } from '@data/state';
import { view } from '@data/view';

const error = view(errorState);

/**
 * Gives the message of the error that was reported last and the way back out of it. The message
 * stays until the next error, so the view can be left and reached again without losing it.
 */
export function useError() {
    return { ...error, showTranslation };
}
