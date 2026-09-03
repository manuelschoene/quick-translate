import type { DeepReadonly } from 'vue';

/**
 * Filters a list of items based on a search string and a filter function.
 *
 * @param search The search string to filter the items by.
 * @param items The list of items to filter.
 * @param filterFn The function that takes an item and returns a string to be used for filtering.
 * @returns The filtered list of items that match the search string.
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
    const filteredItems = items.filter((item) => filterFn(item).toLowerCase().includes(searchLower));
    return filteredItems;
}
