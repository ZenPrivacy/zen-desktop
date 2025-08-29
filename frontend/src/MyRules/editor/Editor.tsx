import { TextArea } from '@blueprintjs/core';
import React, { useMemo, useRef, useCallback } from 'react';
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
