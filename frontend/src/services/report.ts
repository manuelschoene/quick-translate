import { route } from '@/router';
import { errorState } from '@/state';

/**
 * Brings a failure in front of the user and leaves the error view standing until they go back. Meant
 * for everything that works on what is on screen, where there is nothing left to show once it failed.
 */
export function report(error: unknown): void {
    note(error);
    route('error');
}

/**
 * Takes a failure that leaves the view intact, which is the case for the calls beside the
 * translation: throwing the user out of a translation that is still there would cost more than the
 * failure itself. The message is kept anyway, so the error view shows it when it is reached next.
 */
export function reportQuietly(error: unknown): void {
    note(error);
}

/**
 * Puts the message of a failure into the state, which keeps it around while the window is being worked
 * on.
 */
function note(error: unknown): void {
    errorState.message = message(error);
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
