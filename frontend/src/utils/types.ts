import type { Component } from 'vue';

/**
 * Everything a button of the application is built from. A button carries a text, an icon or both,
 * and `inline` drops the frame it draws around itself so that it can be placed inside a container
 * that already brings the height, the border and the background along. `class`, `classText` and
 * `classIcon` are merged onto the button, onto its text and onto its icon by the call site, which is
 * how the few buttons that need to stretch, truncate, fade or animate are told apart from the rest.
 */
export interface ButtonProps {
    action: () => void;
    title: string;
    text?: string;
    icon?: Component;
    disabled?: boolean;
    inline?: boolean;
    class?: string;
    classText?: string;
    classIcon?: string;
}
