<script lang="ts">
	let {
		checked,
		label,
		size = 18,
		touch = false,
		onclick
	}: {
		checked: boolean;
		label: string;
		size?: 14 | 18;
		/** Wrap the box in a 44px tap target. */
		touch?: boolean;
		onclick: () => void;
	} = $props();
</script>

<button
	type="button"
	aria-label={label}
	aria-pressed={checked}
	class="group flex shrink-0 items-center justify-center p-0 {touch ? 'size-11' : ''}"
	onclick={(e) => {
		e.stopPropagation();
		onclick();
	}}
>
	<span
		class="flex items-center justify-center rounded-[2px] border-[1.5px] text-xs leading-none {checked
			? 'border-green-edge bg-green-check text-green'
			: 'border-muted group-hover:border-green'}"
		style:width="{size}px"
		style:height="{size}px">{checked ? '✓' : ''}</span
	>
</button>
