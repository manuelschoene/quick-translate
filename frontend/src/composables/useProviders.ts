import { changeProvider } from '@data/actions';
import { providerState, requestState } from '@data/state';
import { view } from '@data/view';

const providers = view(providerState);
const { pending } = view(requestState);

/**
 * Gives the components the providers that can be chosen, the one that is in use and the way to
 * switch between them. Switching translates the current text with the new provider and brings the
 * languages along, which the other composables pick up on their own.
 */
export function useProviders() {
    return { ...providers, pending, changeProvider };
}
