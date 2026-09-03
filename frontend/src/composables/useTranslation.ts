import { requestState, translationState } from '@/state';
import { readonlyRefs } from '@lib/reactivity';
import { requestQuietly } from '@services/request';
import { CopyTranslation } from '@wails/go/transport/Adapter';

const translation = readonlyRefs(translationState);
const { pending } = readonlyRefs(requestState);

const exposed = { ...translation, pending, copyTranslation };

/**
 * Gives the components the text that came back for the last translation, whether one is being made
 * right now and the way to take the result over. The text is empty until the first translation was
 * made.
 */
export function useTranslation(): typeof exposed {
    return exposed;
}

/**
 * Writes the current translation to the clipboard. Asked for quietly, because a clipboard that
 * refuses to take the text leaves the translation on screen to try again.
 */
async function copyTranslation(): Promise<void> {
    await requestQuietly(CopyTranslation);
}
