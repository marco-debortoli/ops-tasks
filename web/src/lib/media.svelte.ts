import { MediaQuery } from 'svelte/reactivity';

// Keep these in step with Tailwind's `sm` and `lg` breakpoints.

/** Phone width: modals open as bottom sheets. */
export const phone = new MediaQuery('max-width: 639.98px');

/** Desktop width: the three-column dashboard; below it, the mobile dashboard. */
export const desktop = new MediaQuery('min-width: 1024px');
