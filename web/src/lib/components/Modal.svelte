<script lang="ts">
	import type { Snippet } from 'svelte';
	import { onMount } from 'svelte';

	let {
		labelledby,
		width,
		height,
		onclose,
		children
	}: {
		labelledby: string;
		width: number;
		height: number;
		onclose: () => void;
		children: Snippet;
	} = $props();

	let dialog: HTMLDialogElement;

	// A native modal dialog: focus trapping, stacking and Esc come from the browser.
	onMount(() => {
		dialog.showModal();
		// Chrome can hand focus to the first button despite `autofocus` on the dialog.
		dialog.focus();
		return () => dialog.close();
	});

	// Phone: drag the sheet's handle down to dismiss it.
	let dragFrom: number | null = null;
	let dragBy = $state(0);

	function dragStart(e: PointerEvent) {
		dragFrom = e.clientY;
		(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
	}
	function dragMove(e: PointerEvent) {
		if (dragFrom !== null) dragBy = Math.max(0, e.clientY - dragFrom);
	}
	function dragEnd() {
		if (dragFrom === null) return;
		dragFrom = null;
		if (dragBy > 90) onclose();
		else dragBy = 0;
	}
</script>

<!-- Focus the dialog itself on open, not its first button. -->
<!-- svelte-ignore a11y_autofocus -->
<dialog
	bind:this={dialog}
	autofocus
	tabindex="-1"
	aria-labelledby={labelledby}
	oncancel={(e) => {
		e.preventDefault();
		onclose();
	}}
	onclick={(e) => {
		if (e.target === dialog) onclose();
	}}
	class="sheet m-auto max-h-none w-[min(var(--w),calc(100vw-24px))] max-w-none flex-col overflow-hidden rounded-[4px] border border-line-strong bg-modal p-0 text-base text-text outline-none open:flex h-[min(var(--h),calc(100dvh-24px))]
		max-sm:mx-0 max-sm:mt-auto max-sm:mb-0 max-sm:h-[min(var(--h),calc(100dvh-64px))] max-sm:w-full max-sm:rounded-t-[12px] max-sm:rounded-b-none max-sm:border-x-0 max-sm:border-b-0"
	style:--w="{width}px"
	style:--h="{height}px"
	style:translate={dragBy ? `0 ${dragBy}px` : null}
>
	<!-- Grab handle, phone only. Pointer-only: keyboard users have Esc and the close button. -->
	<div
		class="flex shrink-0 touch-none justify-center pt-2.5 pb-2 sm:hidden"
		aria-hidden="true"
		onpointerdown={dragStart}
		onpointermove={dragMove}
		onpointerup={dragEnd}
		onpointercancel={dragEnd}
	>
		<span class="h-1 w-10 rounded-full bg-line-strong"></span>
	</div>
	{@render children()}
</dialog>
