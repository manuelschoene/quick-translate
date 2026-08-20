import { showError } from '@data/navigation';
import { errorState } from '@data/state';

/**
 * Takes a call that failed and brings it in front of the user. Both sources end up here, a call that
 * was rejected and the error event of the shortcut, so both are shown the same way.
 *
 * The message also goes to the console, which keeps it around while the window is being worked on.
 */
export function report(error: unknown): void {
    const text = message(error);

    console.error('Quick Translate:', text);

    errorState.message = text;
    showError();
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
