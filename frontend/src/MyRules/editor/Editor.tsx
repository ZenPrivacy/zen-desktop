import { TextArea } from '@blueprintjs/core';
import React, { useMemo, useRef, useCallback, useLayoutEffect } from 'react';
import { useWraps } from './useWraps';

type Props = {
  value: string;
  placeholder: string;
  disabled: boolean;
  onChange: (next: string) => void;
  lines: string[];
};

export function MyRulesEditor({ value, placeholder, disabled, onChange, lines }: Props) {
  const textAreaRef = useRef<HTMLTextAreaElement>(null);
  const lineNumbersRef = useRef<HTMLDivElement>(null);
  const mirrorRef = useRef<HTMLDivElement>(null);

  const { wraps } = useWraps({ value, lines, textAreaRef, mirrorRef });

  const applyLineNumberMetrics = useCallback(() => {
    const ta = textAreaRef.current;
    const ln = lineNumbersRef.current;
    if (!ta || !ln) return;
    const cs = window.getComputedStyle(ta);

    ln.style.paddingTop = cs.paddingTop;
    ln.style.paddingBottom = cs.paddingBottom;
    ln.style.setProperty('--lnh', cs.lineHeight);
  }, []);

  useLayoutEffect(() => {
    applyLineNumberMetrics();
  }, [applyLineNumberMetrics, value, lines, wraps]);

  useLayoutEffect(() => {
    const ta = textAreaRef.current;
    if (!ta) return;
    const ro = new ResizeObserver(() => applyLineNumberMetrics());
    ro.observe(ta);
    return () => ro.disconnect();
  }, [applyLineNumberMetrics]);

  const syncScroll = useCallback(() => {
    if (lineNumbersRef.current && textAreaRef.current) {
      lineNumbersRef.current.scrollTop = textAreaRef.current.scrollTop;
    }
  }, []);

  const lineNumberItems = useMemo(() => {
    const items: React.ReactNode[] = [];
    for (let i = 0; i < lines.length; i++) {
      const w = wraps[i] ?? 1;
      items.push(
        <div key={`ln-${i}-0`} className="line-number">
          {i + 1}
        </div>,
      );
      for (let k = 1; k < w; k++) {
        items.push(<div key={`ln-${i}-${k}`} className="line-number line-number--cont" />);
      }
    }
    return items;
  }, [lines.length, wraps]);

  return (
    <div className="my-rules__editor">
      <div ref={lineNumbersRef} className="my-rules__line-numbers">
        {lineNumberItems}
      </div>

      <TextArea
        inputRef={textAreaRef}
        fill
        placeholder={placeholder}
        className="my-rules__textarea"
        value={value}
        disabled={disabled}
        onChange={(e) => onChange(e.target.value)}
        onScroll={syncScroll}
      />

      <div ref={mirrorRef} className="my-rules__mirror" aria-hidden>
        <div className="my-rules__mirror-lines" />
      </div>
    </div>
  );
}
