/**
 * A language a provider offers, in the shape the frontend works with. The backend delivers the same
 * information with Go field names, which the data layer maps away in `apply.ts`.
 */
export interface Language {
    tag: string;
    name: string;
    source: boolean;
    target: boolean;
    stable: boolean;
}

/**
 * The views the window can show. The data layer works with the name alone and leaves it to
 * `views/router.ts` to say which component belongs to it.
 */
export type ViewName = 'translation' | 'error' | 'languageSelection';
