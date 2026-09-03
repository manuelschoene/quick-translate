import { onMounted, onUnmounted } from 'vue';

/**
 * Runs the given action whenever the key is pressed, for as long as the view that asks for it is on
 * screen. The listener sits on the window and not on an element, so it also fires while the search
 * box holds the focus, which is where the focus is in every view that has one.
 *
 * Meant for the keys that belong to a whole view, `Escape` above all. A key that belongs to a single
 * element is better handled on that element, which is why `Enter` in the search box is not done here.
 */
export function onKey(key: string, action: () => void): void {
    const handle = (event: KeyboardEvent): void => {
        if (event.key !== key) return;

        event.preventDefault();
        action();
    };

    onMounted(() => {
        window.addEventListener('keydown', handle);
    });
    onUnmounted(() => {
        window.removeEventListener('keydown', handle);
    });
}
