import { errorState } from '@/state';
import { readonlyRefs } from '@lib/reactivity';

const exposed = readonlyRefs(errorState);

/**
 * Gives the components the message of the failure that was reported last.
 */
export function useError(): typeof exposed {
    return exposed;
}
