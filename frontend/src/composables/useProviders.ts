import { providerState } from '@/state';
import { readonlyRefs } from '@lib/reactivity';
import { Globe } from '@lucide/vue';
import { request } from '@services/request';
import { applyFull } from '@services/wire';
import { ChangeProvider } from '@wails/go/transport/Adapter';
import { computed, type Component } from 'vue';
import { DeepLIcon } from 'vue3-simple-icons';

/**
 * A translation service in the shape the components show it: the backend only knows the slug, the
 * icon and the label are what the frontend brings along.
 */
interface Provider {
    icon: Component;
    label: string;
    slug: string;
}

/**
 * The providers this frontend has an icon and a label for. A backend that knows more of them than
 * this list does still works, its providers are shown with the generic icon and their slug.
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
 * The provider that is in use. Until the backend has answered it falls back to the first registered
 * provider and only then to the first one this file knows about, so there is always something to
 * display.
 */
const provider = computed(
    () => registered.value.find((p) => p.slug === current.value) ?? registered.value.at(0) ?? known[0],
);

/**
 * The providers that can be switched to, which is every registered one except the one in use.
 * Filtering against the resolved provider rather than against the raw slug matters while the backend
 * has not answered yet: the slug is still empty then, and comparing against it would report the only
 * registered provider as an alternative to itself.
 */
const alternatives = computed(() => registered.value.filter((p) => p.slug !== provider.value.slug));

const exposed = { provider, alternatives, changeProvider };

/**
 * Gives the components the provider that is in use, the ones that can be switched to and the way to
 * switch. Switching translates the current text with the new provider and brings the languages
 * along, which the other composables pick up on their own.
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
 * Brings a slug into the shape the components show it in. A slug this frontend has no entry for
 * keeps working and is shown with a generic icon and the slug itself, which is better than dropping
 * a provider the backend offers without a word.
 */
function describe(slug: string): Provider {
    return known.find((p) => p.slug === slug) ?? { icon: Globe, label: slug, slug };
}
