import { Button, Tooltip, Spinner, ButtonGroup } from '@blueprintjs/core';
import { useEffect, useState, useRef, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { useDebouncedCallback } from 'use-debounce';

// CodeMirror 6 Core & UI
import { EditorView, keymap, drawSelection, highlightActiveLine, scrollPastEnd, lineNumbers } from "@codemirror/view";
import { EditorState, Compartment } from "@codemirror/state";
import { defaultKeymap, history, historyKeymap } from "@codemirror/commands";
import { 
  search, 
  searchKeymap, 
  openSearchPanel, 
  findNext, 
  replaceNext, 
  replaceAll 
} from "@codemirror/search";
import { oneDark } from "@codemirror/theme-one-dark";

// Language & Highlighting
import { StreamLanguage, syntaxHighlighting, defaultHighlightStyle } from "@codemirror/language";
import { tags as t } from "@lezer/highlight";

import './index.css';
import { GetRules, SetRules } from '../../wailsjs/go/cfg/Config';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
import { useProxyState } from '../context/ProxyStateContext';

const HELP_URL = 'https://github.com/ZenPrivacy/zen-desktop/blob/master/docs/external/how-to-rules.md';

/**
 * Custom Adblock Syntax Highlighter
 */
const adblockLanguage = StreamLanguage.define({
  token(stream) {
    if (stream.sol() && stream.peek() === "!") {
      stream.skipToEnd();
      return "comment";
    }
    if (stream.match(/^@@/)) return "keyword";
    if (stream.match(/^\|\|/)) return "operator";
    stream.next();
    return null;
  },
  tokenTable: {
    comment: t.comment,
    keyword: t.keyword,
    operator: t.operator
  }
});

export function Rules() {
  const { t: translate } = useTranslation();
  const { isProxyRunning } = useProxyState();
  const [initialRules, setInitialRules] = useState<string | null>(null);
  const editorRef = useRef<HTMLDivElement>(null);
  const viewRef = useRef<EditorView | null>(null);
  const readOnlyCompartment = useMemo(() => new Compartment(), []);

  useEffect(() => {
    (async () => {
      const filters = await GetRules();
      setInitialRules(filters?.join('\n') ?? '');
    })();
  }, []);

  const setFilters = useDebouncedCallback(async (rules: string) => {
    await SetRules(
      rules.split('\n').map((f) => f.trim()).filter((f) => f.length > 0)
    );
  }, 500);

  /**
   * Format Logic: Removes duplicates and sorts rules.
   * Comments (!) are kept at the top.
   */
  const handleFormat = () => {
    if (!viewRef.current) return;
    const currentDoc = viewRef.current.state.doc.toString();
    const lines = currentDoc.split('\n')
      .map(line => line.trim())
      .filter(line => line.length > 0);

    const formatted = Array.from(new Set(lines))
      .sort((a, b) => {
        const aIsComment = a.startsWith('!');
        const bIsComment = b.startsWith('!');
        if (aIsComment && !bIsComment) return -1;
        if (!aIsComment && bIsComment) return 1;
        return a.localeCompare(b);
      })
      .join('\n');

    viewRef.current.dispatch({
      changes: { from: 0, to: viewRef.current.state.doc.length, insert: formatted }
    });
  };

  useEffect(() => {
    if (initialRules !== null && editorRef.current && !viewRef.current) {
      const state = EditorState.create({
        doc: initialRules,
        extensions: [
          lineNumbers(),
          history(),
          drawSelection(),
          highlightActiveLine(),
          scrollPastEnd(),
          // FIX: Changed 'topMost' to 'top' to fix TS2345
          search({ top: true }), 
          keymap.of([...defaultKeymap, ...historyKeymap, ...searchKeymap]),
          oneDark,
          adblockLanguage,
          syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
          readOnlyCompartment.of(EditorState.readOnly.of(isProxyRunning)),
          EditorView.lineWrapping,
          EditorView.updateListener.of((update) => {
            if (update.docChanged) {
              setFilters(update.state.doc.toString());
            }
          }),
        ],
      });
      viewRef.current = new EditorView({ state, parent: editorRef.current });
    }
    return () => {
      viewRef.current?.destroy();
      viewRef.current = null;
    };
  }, [initialRules]);

  useEffect(() => {
    if (viewRef.current) {
      viewRef.current.dispatch({
        effects: readOnlyCompartment.reconfigure(EditorState.readOnly.of(isProxyRunning))
      });
    }
  }, [isProxyRunning, readOnlyCompartment]);

  if (initialRules === null) {
    return <div className="rules--loading"><Spinner /></div>;
  }

  return (
    <div className="rules">
      <div className="rules__header">
        <ButtonGroup>
          <Button variant="outlined" icon="help" onClick={() => BrowserOpenURL(HELP_URL)}>
            {translate('rules.help')}
          </Button>
          <Button 
            variant="outlined" 
            icon="clean" 
            disabled={isProxyRunning} 
            onClick={handleFormat}
            title={translate('rules.formatTooltip') || 'Sort and remove duplicates'}
          >
            {translate('rules.format') || 'Format'}
          </Button>
        </ButtonGroup>
        <div className="rules__hint">Press Ctrl+F to Search</div>
      </div>
      <Tooltip
        content={translate('common.stopProxyToEditRules') as string}
        disabled={!isProxyRunning}
        className="rules__tooltip"
        placement="top"
      >
        <div 
          ref={editorRef} 
          className={`rules__editor-container ${isProxyRunning ? 'is-disabled' : ''}`} 
        />
      </Tooltip>
    </div>
  );
}
