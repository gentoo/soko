// Helpers for building HTML fragments out of untrusted values.

const HTML_ESCAPES = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;',
    '{': '&#123;',
    '}': '&#125;',
};

export function escapeHtml(value) {
    return String(value === null || value === undefined ? '' : value)
        .replace(/[&<>"'{}]/g, function (char) { return HTML_ESCAPES[char]; });
}

export function highlightHtml(value, query) {
    const escaped = escapeHtml(value);

    const terms = String(query === null || query === undefined ? '' : query)
        .split(' ')
        .filter(function (term) { return term.trim() !== ''; })
        .map(function (term) {
            return escapeHtml(term).replace(/[-[\]{}()*+?.,\\^$|#\s]/g, '\\$&');
        })
        .sort(function (a, b) { return b.length - a.length; });

    if (terms.length === 0) {
        return escaped;
    }

    return escaped.replace(
        new RegExp('(' + terms.join('|') + ')', 'gi'),
        '<strong>$1</strong>'
    );
}
