import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

/**
 * Joins Tailwind classes and resolves the ones that collide, so a class handed in by the call site
 * wins over the one a component brings along instead of both ending up on the element.
 */
export function cn(...inputs: ClassValue[]): string {
    return twMerge(clsx(inputs));
}
