import { report } from '@data/report';
import { requestState } from '@data/state';

/**
 * How many calls are on their way. Counted instead of a plain flag, because the initial load starts
 * three calls at once and the first one to return would otherwise clear the pending state for the
 * other two.
 */
let running = 0;

/**
 * Runs a call into the backend, holds the pending state while it is on its way and hands the result
 * to the given function. A call that fails is reported instead of thrown, so nothing above this
 * layer has to deal with it and the state simply stays as it was.
 *
 * The function is optional for the calls that return nothing.
 */
export async function request<T>(call: () => Promise<T>, apply?: (result: T) => void): Promise<void> {
    begin();

    try {
        const result = await call();
        apply?.(result);
    } catch (error) {
        report(error);
    } finally {
        finish();
    }
}

/**
 * Notes a call that is on its way. Public because a translation by the shortcut is started in the
 * backend, so the frontend only sees its two ends as events.
 */
export function begin(): void {
    running++;
    requestState.pending = true;
}

/**
 * Notes a call that is through. Does not count below zero, so an end without a beginning can not
 * leave the pending state stuck.
 */
export function finish(): void {
    running = Math.max(0, running - 1);
    requestState.pending = running > 0;
}
