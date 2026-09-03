import { errorState } from '@/state';
import { readonlyRefs } from '@lib/reactivity';

const error = readonlyRefs(errorState);

/**
 * Gives the message of the error that was reported last. The message stays until the next error, so
 * the view can be left and reached again without losing it.
 */
export function useError(): typeof error {
    return error;
}
