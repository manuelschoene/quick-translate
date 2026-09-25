import { computed, onUnmounted, ref, type ComputedRef, type Ref } from 'vue';

export type Outcome = 'none' | 'success' | 'failure';

interface Feedback {
    outcome: Readonly<Ref<Outcome>>;
    iconClass: ComputedRef<string | undefined>;
    show: (succeeded: boolean) => void;
}

/**
 * How long the answer stands before it fades, and how long the fade takes. The fade has to match
 * `--animate-check-out` in `style.css`, otherwise the icon is swapped back while it is still there
 * or stays invisible for a moment.
 */
const holdDuration = 1500;
const fadeDuration = 500;

/**
 * Answers on the button that was pressed, for the calls that are reported quietly: they leave the
 * view as it is, so without this a failed copy or a window that refuses to close would look exactly
 * like one that worked.
 *
 * The component decides which icon each outcome stands for, this only says which one is due and how
 * it is colored and animated. Asks for the setup scope, like every composable, so a component that
 * is left while the answer stands does not run its timers into nothing.
 */
export function feedback(): Feedback {
    const outcome = ref<Outcome>('none');
    const fading = ref(false);

    let timeout: ReturnType<typeof setTimeout>;

    const iconClass = computed(() => {
        if (outcome.value === 'none') {
            return undefined;
        }

        const color = outcome.value === 'success' ? 'text-(--text-success)' : 'text-(--text-error)';

        return `${fading.value ? 'animate-check-out' : 'animate-check'} ${color}`;
    });

    /**
     * Puts the answer up and takes it down again. Pressing the button while an answer still stands
     * starts it over, so the second press is confirmed and not swallowed by the first.
     */
    function show(succeeded: boolean): void {
        clearTimeout(timeout);

        outcome.value = succeeded ? 'success' : 'failure';
        fading.value = false;

        timeout = setTimeout(() => {
            fading.value = true;

            timeout = setTimeout(() => {
                outcome.value = 'none';
                fading.value = false;
            }, fadeDuration);
        }, holdDuration);
    }

    onUnmounted(() => {
        clearTimeout(timeout);
    });

    return { outcome, iconClass, show };
}
