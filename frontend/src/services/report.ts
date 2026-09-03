import { route } from '@/router';
import { errorState } from '@/state';

/**
 * Takes a call that failed, brings it in front of the user and leaves the error view standing until
 * the user goes back. Meant for everything that works on what is on screen, where there is nothing
 * left to show once it failed.
 */
export function report(error: unknown): void {
    note(error);
    route('error');
}

/**
 * Takes a call that failed without leaving the view the user is on. Meant for the calls beside the
 * translation, copying and dismissing, which leave the state intact: throwing the user out of a
 * translation that is still there would cost more than the failure itself.
 *
 * The message is kept anyway, so the error view shows it when it is reached the next time.
 */
export function reportQuietly(error: unknown): void {
    note(error);
}

/**
 * Puts the message of a failure into the state and onto the console, which keeps it around while
 * the window is being worked on. Both sources end up here, a call that was rejected and the error
 * event of the shortcut, so both are shown the same way.
 */
function note(error: unknown): void {
    const text = message(error);

    console.error('Quick Translate: ', text);

    errorState.message = text;
}

/**
 * Turns whatever a rejected call carries into a message. Wails rejects with the string of the Go
 * error and not with an Error, so the plain string is the regular case here.
 */
function message(error: unknown): string {
    if (typeof error === 'string') {
        return error;
    }

    if (error instanceof Error) {
        return error.message;
    }

    return String(error);
}
