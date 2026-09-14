import { requestState, translationState } from '@/state';
import { CopyTranslation } from '@bind/transport/adapter';
import { readonlyRefs } from '@lib/reactivity';
import { requestQuietly } from '@services/request';

const translation = readonlyRefs(translationState);
const { pending } = readonlyRefs(requestState);

const exposed = { ...translation, pending, copyTranslation };

/**
 * Gives the components the text that came back for the last translation, whether one is being made
 * right now and the way to take the result over.
 */
export function useTranslation(): typeof exposed {
    return exposed;
}

/**
 * Writes the current translation to the clipboard and answers whether it worked. Asked for quietly,
 * because a clipboard that refuses to take the text leaves the translation on screen to try again.
 */
async function copyTranslation(): Promise<boolean> {
    return await requestQuietly(CopyTranslation);
}
