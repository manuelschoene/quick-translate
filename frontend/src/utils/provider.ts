import { useProviders } from '@/composables/useProviders';
import { computed, type Component, type ComputedRef } from 'vue';
import { DeepLIcon } from 'vue3-simple-icons';

interface Provider {
    icon: Component;
    label: string;
    slug: string;
}

const providers = [
    {
        icon: DeepLIcon,
        label: 'DeepL',
        slug: 'deepl',
    },
];

/**
 * Returns the providers the backend has registered, mapped to the icon and label they are shown
 * with. Slugs this list has no entry for are dropped, so a backend that knows more providers than
 * the frontend does never renders a blank button.
 */
function registeredProviders(): ComputedRef<Provider[]> {
    const { providers: registered } = useProviders();

    return computed(() =>
        registered.value.map((slug) => providers.find((p) => p.slug === slug)).filter((p) => p !== undefined),
    );
}

/**
 * Returns the currently selected provider's icon and label. Until the backend has answered, and for
 * a slug the frontend does not know, it falls back to the first registered provider and only then to
 * the first one this file knows about, so there is always something to display.
 */
export function currentProvider(): ComputedRef<Provider> {
    const { current } = useProviders();
    const registered = registeredProviders();

    return computed(() => {
        const provider = registered.value.find((p) => p.slug === current.value);
        return provider ?? registered.value[0] ?? providers[0];
    });
}

/**
 * Returns the registered providers that can be switched to, which is every registered one except
 * the current one. Filtering against the resolved current provider rather than against the raw slug
 * matters while the backend has not answered yet: the slug is still empty then, and comparing
 * against it would report the only registered provider as an alternative to itself.
 */
export function alternativeProviders(): ComputedRef<Provider[]> {
    const current = currentProvider();
    const registered = registeredProviders();

    return computed(() => registered.value.filter((p) => p.slug !== current.value.slug));
}
