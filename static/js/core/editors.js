// CodeMirror 6 editors with a "Warm Paper Workbench" theme.
//
// The theme is expressed entirely with the site's CSS custom properties, so a
// single editor instance follows the active theme without reconfiguration.
// Packages are resolved through the import map declared in index.html.
//
// CM6 is imported dynamically: if the CDN is unavailable the rest of the
// workbench keeps working and the inputs simply remain plain textareas.

let cm = null;
let loading = null;

function loadCodeMirror() {
    if (cm) return Promise.resolve(cm);
    if (loading) return loading;

    loading = Promise.all([
        import('codemirror'),
        import('@codemirror/state'),
        import('@codemirror/language'),
        import('@codemirror/lang-json'),
        import('@codemirror/lang-yaml'),
        import('@codemirror/lang-sql'),
        import('@lezer/highlight')
    ]).then(([core, state, language, langJson, langYaml, langSql, lezer]) => {
        const { EditorView, basicSetup } = core;
        const { EditorState } = state;
        const { HighlightStyle, syntaxHighlighting } = language;
        const { tags } = lezer;

        const theme = EditorView.theme({
            '&': {
                color: 'var(--text-primary)',
                backgroundColor: 'color-mix(in srgb, var(--text-primary) 5%, transparent)',
                border: '1px solid var(--card-border)',
                borderRadius: 'var(--radius-md)',
                fontSize: '13px',
                overflow: 'hidden'
            },
            '&.cm-focused': {
                outline: 'none',
                borderColor: 'color-mix(in srgb, var(--color-primary) 65%, transparent)',
                boxShadow: '0 0 0 3px color-mix(in srgb, var(--color-primary) 18%, transparent)'
            },
            '.cm-scroller': {
                fontFamily: 'var(--font-mono)',
                lineHeight: '1.6'
            },
            '.cm-content': {
                caretColor: 'var(--color-primary)',
                padding: '10px 0'
            },
            '.cm-gutters': {
                backgroundColor: 'color-mix(in srgb, var(--text-primary) 4%, transparent)',
                color: 'var(--text-muted)',
                borderRight: '1px solid var(--card-border)'
            },
            '.cm-activeLine': {
                backgroundColor: 'color-mix(in srgb, var(--text-primary) 3%, transparent)'
            },
            '.cm-activeLineGutter': {
                backgroundColor: 'color-mix(in srgb, var(--text-primary) 3%, transparent)'
            },
            '.cm-cursor, .cm-dropCursor': {
                borderLeftColor: 'var(--color-primary)'
            },
            '&.cm-focused .cm-selectionBackground, .cm-selectionBackground, .cm-content ::selection': {
                backgroundColor: 'color-mix(in srgb, var(--color-primary) 22%, transparent)'
            },
            '.cm-panels': {
                backgroundColor: 'var(--bg-secondary)',
                color: 'var(--text-primary)'
            }
        }, { dark: true });

        const highlight = HighlightStyle.define([
            { tag: tags.propertyName, color: 'var(--color-accent)' },
            { tag: [tags.string, tags.special(tags.string)], color: 'var(--success)' },
            { tag: [tags.number, tags.bool, tags.null], color: 'var(--color-pink)' },
            { tag: [tags.comment, tags.lineComment, tags.blockComment], color: 'var(--text-muted)', fontStyle: 'italic' },
            { tag: [tags.keyword, tags.operatorKeyword, tags.controlKeyword], color: 'var(--color-primary)' },
            { tag: [tags.function(tags.variableName), tags.labelName], color: 'var(--color-accent)' },
            { tag: [tags.typeName, tags.className, tags.definition(tags.typeName)], color: 'var(--color-pink)' },
            { tag: tags.tagName, color: 'var(--color-primary)' },
            { tag: tags.attributeName, color: 'var(--color-accent)' }
        ]);

        cm = {
            EditorView,
            EditorState,
            basicSetup,
            theme,
            highlighting: syntaxHighlighting(highlight),
            languages: {
                'application/json': langJson.json,
                'text/x-yaml': langYaml.yaml,
                'application/yaml': langYaml.yaml,
                'text/x-sql': langSql.sql
            }
        };
        return cm;
    }).catch((err) => {
        console.error('CodeMirror failed to load; inputs stay as plain textareas.', err);
        loading = null;
        return null;
    });

    return loading;
}

// Named, editable editors mounted on tool inputs (json / yaml / sql).
const editors = new Map();
// Read-only result viewers created after an HTMX swap.
const resultEditors = new Set();

function buildExtensions(instance, mode, readOnly) {
    const extensions = [
        instance.basicSetup,
        instance.theme,
        instance.highlighting,
        instance.EditorView.lineWrapping
    ];
    const language = instance.languages[mode];
    if (language) extensions.push(language());
    if (readOnly) {
        extensions.push(instance.EditorState.readOnly.of(true));
        extensions.push(instance.EditorView.editable.of(false));
    }
    return extensions;
}

// Replace a <textarea> with a CodeMirror 6 instance and return a small handle.
function mount(instance, textarea, mode, readOnly) {
    const extensions = buildExtensions(instance, mode, readOnly);
    // Keep the underlying textarea in sync: HTMX hx-include and the tools read
    // its value, so the editor must mirror every document change back into it.
    extensions.push(instance.EditorView.updateListener.of((update) => {
        if (update.docChanged) textarea.value = update.state.doc.toString();
    }));

    const state = instance.EditorState.create({
        doc: textarea.value,
        extensions
    });
    const view = new instance.EditorView({ state, parent: textarea.parentNode });
    textarea.style.display = 'none';
    textarea.dataset.cmInitialized = 'true';

    return {
        view,
        getValue: () => view.state.doc.toString(),
        setValue: (value) => view.dispatch({
            changes: { from: 0, to: view.state.doc.length, insert: value }
        }),
        refresh: () => view.requestMeasure(),
        focus: () => view.focus()
    };
}

export async function mountEditor(name, textareaId, mode) {
    const textarea = document.getElementById(textareaId);
    if (!textarea || textarea.dataset.cmInitialized === 'true') return null;

    const instance = await loadCodeMirror();
    if (!instance) return null;

    try {
        const handle = mount(instance, textarea, mode, false);
        editors.set(name, handle);
        return handle;
    } catch (err) {
        console.error(`Could not mount the ${name} editor:`, err);
        return null;
    }
}

export function getEditor(name) {
    return editors.get(name) || null;
}

export function setEditorValue(name, value) {
    const handle = editors.get(name);
    if (handle) handle.setValue(value);
}

export function refreshEditors() {
    for (const handle of editors.values()) handle.refresh();
}

// Idempotently upgrade any read-only result <textarea> inside a container.
export async function initResultEditors(root = document) {
    const textareas = Array.from(root.querySelectorAll('textarea.result-text'))
        .filter((textarea) => textarea.dataset.cmInitialized !== 'true');
    if (!textareas.length) return;

    const instance = await loadCodeMirror();
    if (!instance) return;

    textareas.forEach((textarea) => {
        if (textarea.dataset.cmInitialized === 'true') return;
        try {
            const mode = textarea.getAttribute('data-mode') || 'text/x-yaml';
            resultEditors.add(mount(instance, textarea, mode, true));
        } catch (err) {
            console.error('Could not mount a result editor:', err);
        }
    });
}
