import { report, reportQuietly } from '@services/report';

/**
 * Runs a call that works on what is shown and hands the result to the given function. A failure is reported
 * instead of thrown, so no caller needs try/catch. Whether a translation is pending comes from its events, not
 * from here.
 */
export async function request<T>(call: () => Promise<T>, apply?: (result: T) => void): Promise<void> {
    try {
        const result = await call();
        apply?.(result);
    } catch (error) {
        report(error);
    }
}

/**
 * Runs a call beside the translation, such as copying or hiding, which keeps the user out of the error view
 * when it fails. Answers whether it went through, so the pressed button can show the outcome.
 */
export async function requestQuietly<T>(call: () => Promise<T>, apply?: (result: T) => void): Promise<boolean> {
    try {
        const result = await call();
        apply?.(result);

        return true;
    } catch (error) {
        reportQuietly(error);

        return false;
    }
}
