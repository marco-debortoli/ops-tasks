import DOMPurify from 'dompurify';
import { marked } from 'marked';

// Links in a rendered description open in a new tab, so following one doesn't leave the app.
DOMPurify.addHook('afterSanitizeAttributes', (node) => {
	if (node.tagName === 'A' && node.hasAttribute('href')) {
		node.setAttribute('target', '_blank');
		node.setAttribute('rel', 'noopener noreferrer');
	}
});

/** Render markdown to HTML that is safe to insert with {@html}. */
export function renderMarkdown(src: string): string {
	const html = marked.parse(src, { async: false, gfm: true, breaks: true });
	return DOMPurify.sanitize(html);
}
