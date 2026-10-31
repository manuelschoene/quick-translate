import { providerState } from '@/state';
import { ProviderID } from '@bind/models';
import type { Info, State } from '@bind/provider';
import { Current, Providers, Select } from '@bind/provider/service';
import { on } from '@lib/events';
import { readonlyRefs } from '@lib/reactivity';
import { Globe } from '@lucide/vue';
import { request } from '@services/request';
import { retranslate } from '@use/useTranslation';
import { computed, type Component, type DeepReadonly } from 'vue';
import { DeepLIcon } from 'vue3-simple-icons';

interface Provider {
    icon: Component;
    label: string;
    id: ProviderID;
}

/**
 * The icons this frontend has for a provider. The name comes from the backend.
 */
const icons: Partial<Record<ProviderID, Component>> = {
    [ProviderID.DeepL]: DeepLIcon,
};

/**
 * Shown while the backend has not answered.
 */
const placeholder: Provider = { icon: Globe, label: '', id: ProviderID.$zero };

const { current, providers } = readonlyRefs(providerState);

/**
 * The configured providers, in the shape they are shown in.
 */
const registered = computed(() => providers.value.map(describe));

/**
 * The provider in use. Falls back to the first configured one, then to the placeholder.
 */
const provider = computed(
    () => registered.value.find((p) => p.id === current.value?.id) ?? registered.value.at(0) ?? placeholder,
);

/**
 * The providers that can be switched to: every configured one except the one in use.
 */
const alternatives = computed(() => registered.value.filter((p) => p.id !== provider.value.id));

const exposed = { provider, alternatives, changeProvider };

/**
 * Gives the components the provider in use, the ones to switch to and the way to switch.
 */
export function useProviders(): typeof exposed {
    return exposed;
}

/**
 * Listens to provider switches. Returns the unsubscribe.
 */
export function listen(): () => void {
    return on('provider:changed', apply);
}

/**
 * Fetches the providers, for the first render. Fails while no provider is configured, which the error view
 * then shows.
 */
export async function load(): Promise<void> {
    const [all, selected] = await Promise.all([Providers(), Current()]);

    apply({ current: selected, providers: all });
}

/**
 * Switches to the given provider and translates the shown text with it. The languages follow as events.
 */
async function changeProvider(id: ProviderID): Promise<void> {
    await request(async () => {
        await Select(id);
        await retranslate();
    });
}

/**
 * Takes the providers over.
 */
function apply(state: State): void {
    providerState.current = state.current;
    providerState.providers = state.providers;
}

/**
 * Brings a provider into the shape it is shown in. One without an icon gets a generic one.
 */
function describe(info: DeepReadonly<Info>): Provider {
    return { icon: icons[info.id] ?? Globe, label: info.name, id: info.id };
}
