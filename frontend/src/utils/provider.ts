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
 * Returns the currently selected provider's icon and label. If no provider is selected, it returns the default provider (the first one in the list).
 */
export function currentProvider(): ComputedRef<Provider> {
    const { current } = useProviders();

    return computed(() => {
        const provider = providers.find((p) => p.slug === current.value);
        return provider ?? providers[0];
    });
}

/**
 * Returns a list of providers excluding the currently selected one. This is useful for displaying alternative provider options in a dropdown or selection menu.
 */
export function alternativeProviders(): ComputedRef<Provider[]> {
    const { current } = useProviders();

    return computed(() => {
        return providers.filter((p) => p.slug !== current.value);
    });
}
