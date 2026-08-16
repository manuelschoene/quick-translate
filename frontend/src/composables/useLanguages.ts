import { changeSource, changeTarget, switchLanguages } from '@data/actions';
import { languageState, requestState } from '@data/state';
import { view } from '@data/view';

const languages = view(languageState);
const { pending } = view(requestState);

/**
 * Gives the components the languages of the current provider, the ones that are chosen and the ways
 * to change them. Every change translates the current text again, so the components only pick a
 * language and read the new state.
 *
 * Language detection is not part of the source languages, `detection` tells whether it can be
 * offered on top of them and `detectedSource` holds what the provider made of the last text.
 */
export function useLanguages() {
    return { ...languages, pending, changeSource, changeTarget, switchLanguages };
}
