import { providerState } from '@/state';
import { ChangeProvider } from '@bind/transport/adapter';
import { readonlyRefs } from '@lib/reactivity';
import { Globe } from '@lucide/vue';
import { request } from '@services/request';
import { applyFull } from '@services/wire';
import { computed, type Component } from 'vue';
import { DeepLIcon } from 'vue3-simple-icons';

interface Provider {
    icon: Component;
    label: string;
    slug: string;
}

/**
 * The providers this frontend has an icon and a label for. The backend only ever names a slug, so
 * this is the whole of what the application knows about how a provider looks.
 */
const known: Provider[] = [
    {
        icon: DeepLIcon,
        label: 'DeepL',
        slug: 'deepl',
    },
];

const { current, providers } = readonlyRefs(providerState);

/**
 * The providers the backend has registered, in the shape they are shown in.
 */
const registered = computed(() => providers.value.map(describe));

/**
 * The provider that is in use. Falls back to the first registered one and then to the first known
 * one, so there is always something to display while the backend has not answered.
 */
const provider = computed(
    () => registered.value.find((p) => p.slug === current.value) ?? registered.value.at(0) ?? known[0],
);

/**
 * The providers that can be switched to, which is every registered one except the one in use.
 *
 * Compared against the resolved provider and not against the raw slug: the slug is still empty until
 * the backend answers, and comparing against it would offer the only provider as an alternative to
 * itself.
 */
const alternatives = computed(() => registered.value.filter((p) => p.slug !== provider.value.slug));

const exposed = { provider, alternatives, changeProvider };

/**
 * Gives the components the provider that is in use, the ones that can be switched to and the way to
 * switch. Switching translates the current text again and brings new languages along.
 */
export function useProviders(): typeof exposed {
    return exposed;
}

/**
 * Switches to the given provider and translates the current text with it.
 */
async function changeProvider(slug: string): Promise<void> {
    await request(() => ChangeProvider(slug), applyFull);
}

/**
 * Brings a slug into the shape the components show it in. A slug without an entry keeps working and
 * is shown with a generic icon and the slug itself, which is better than dropping a provider the
 * backend offers without a word.
 */
function describe(slug: string): Provider {
    return known.find((p) => p.slug === slug) ?? { icon: Globe, label: slug, slug };
}
