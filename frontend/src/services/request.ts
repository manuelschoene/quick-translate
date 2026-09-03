import { requestState } from '@/state';
import { report, reportQuietly } from '@services/report';

/**
 * How many calls that work on what is shown are on their way. Counted instead of a plain flag,
 * because one of them can be started while another one is still running and the first one to return
 * would otherwise clear the pending state for the rest.
 */
let running = 0;

/**
 * Whether the backend is translating on its own after the shortcut was pressed. Held as a flag and
 * not counted, because the frontend only sees the two ends of that translation as events: a second
 * shortcut while the first one is still running, or an end that never arrives, would make a counter
 * drift, while a flag simply describes the last thing that was heard.
 */
let shortcut = false;

/**
 * Runs a call that works on what is shown, holds the pending state while it is on its way and hands
 * the result to the given function. A call that fails is reported instead of thrown, so nothing
 * above this layer has to deal with it and the state simply stays as it was.
 *
 * The function is optional for the calls that return nothing.
 */
export async function request<T>(call: () => Promise<T>, apply?: (result: T) => void): Promise<void> {
    running++;
    settle();

    try {
        const result = await call();
        apply?.(result);
    } catch (error) {
        report(error);
    } finally {
        running--;
        settle();
    }
}

/**
 * Runs a call beside the translation, which neither shows that it is working nor takes the user to
 * the error view when it fails. Used for copying and dismissing: both leave what is on screen as it
 * is, so a failure is worth a message but not the view.
 */
export async function requestQuietly<T>(call: () => Promise<T>, apply?: (result: T) => void): Promise<void> {
    try {
        const result = await call();
        apply?.(result);
    } catch (error) {
        reportQuietly(error);
    }
}

/**
 * Notes that the backend started translating after the shortcut was pressed.
 */
export function beginShortcut(): void {
    shortcut = true;
    settle();
}

/**
 * Notes that the translation the shortcut asked for is through, however it went. Setting the flag
 * back is enough, so an error, a second shortcut or a frontend that was mounted again in `wails dev`
 * can not leave the pending state stuck.
 */
export function endShortcut(): void {
    shortcut = false;
    settle();
}

/**
 * Writes down whether anything is being worked on, which is the case while a call of this frontend
 * or a translation of the shortcut is running.
 */
function settle(): void {
    requestState.pending = running > 0 || shortcut;
}
