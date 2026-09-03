import type { DeepReadonly } from 'vue';

/**
 * Keeps the items whose text contains the search, ignoring case. An empty search gives the list back
 * unchanged, so a caller can hand its input straight through without checking it first.
 */
export function searchFilter<T>(
    search: string,
    items: DeepReadonly<T[]>,
    filterFn: (item: DeepReadonly<T>) => string,
): DeepReadonly<T[]> {
    if (!search) {
        return items;
    }

    const searchLower = search.toLowerCase();

    return items.filter((item) => filterFn(item).toLowerCase().includes(searchLower));
}
