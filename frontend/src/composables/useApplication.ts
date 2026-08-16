import { hide, start, stop } from '@data/actions';
import { requestState } from '@data/state';
import { view } from '@data/view';

const { pending } = view(requestState);

/**
 * Gives the components what belongs to the application as a whole: whether it is waiting for the
 * backend, the ways to bring it up and take it down again and the way to dismiss it.
 *
 * The pending state is shared by every composable, because the backend works through one request at
 * a time.
 */
export function useApplication() {
    return { pending, start, stop, hide };
}
