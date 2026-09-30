/**
 * Autosave for modal fields. `save` sends a change immediately (buttons, selects);
 * `queue` batches typing and sends it after a pause; call `flush` before closing.
 */
export function autosave<P extends object>(send: (patch: P) => Promise<unknown>) {
	const status = $state({ savedAt: null as Date | null, error: '' });
	let pending: Partial<P> = {};
	let timer: ReturnType<typeof setTimeout> | undefined;

	async function save(patch: P) {
		try {
			await send(patch);
			status.savedAt = new Date();
			status.error = '';
		} catch (e) {
			status.error = e instanceof Error ? e.message : String(e);
		}
	}

	function queue(patch: Partial<P>) {
		Object.assign(pending, patch);
		clearTimeout(timer);
		timer = setTimeout(flush, 600);
	}

	function flush(): Promise<void> {
		clearTimeout(timer);
		if (Object.keys(pending).length === 0) return Promise.resolve();
		const patch = pending as P;
		pending = {};
		return save(patch);
	}

	return { status, save, queue, flush };
}
