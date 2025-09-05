import { useCallback, useEffect, useLayoutEffect, useState } from 'react';

type Args = {
  value: string;
  lines: string[];
  textAreaRef: React.RefObject<HTMLTextAreaElement>;
  mirrorRef: React.RefObject<HTMLDivElement>;
};

export function useWraps({ value, lines, textAreaRef, mirrorRef }: Args) {
  const [wraps, setWraps] = useState<number[]>([]);

  const computeLineHeightPx = useCallback((): number => {
    const ta = textAreaRef.current;
    if (!ta) return 20;
    const cs = window.getComputedStyle(ta);
    const lh = cs.lineHeight;
    if (!lh || lh === 'normal') {
      const fs = parseFloat(cs.fontSize || '16');
      return fs * 1.2;
    }
    return parseFloat(lh);
  }, [textAreaRef]);

  const recalcWraps = useCallback(() => {
    const ta = textAreaRef.current;
    const mirror = mirrorRef.current;
    if (!ta || !mirror) return;

    const cs = window.getComputedStyle(ta);
    const innerWidth = ta.clientWidth - parseFloat(cs.paddingLeft) - parseFloat(cs.paddingRight);

    mirror.style.width = `${Math.max(0, innerWidth)}px`;
    mirror.style.fontFamily = cs.fontFamily;
    mirror.style.fontSize = cs.fontSize;
    mirror.style.lineHeight = cs.lineHeight;
    mirror.style.letterSpacing = cs.letterSpacing;
    mirror.style.wordSpacing = cs.wordSpacing;
    mirror.style.fontWeight = cs.fontWeight;
    mirror.style.boxSizing = cs.boxSizing;
    mirror.style.paddingTop = cs.paddingTop;
    mirror.style.paddingBottom = cs.paddingBottom;

    const mirrorLines = mirror.querySelector('.my-rules__mirror-lines') as HTMLDivElement | null;
    if (!mirrorLines) return;
    mirrorLines.innerHTML = '';

    const lineHeightPx = computeLineHeightPx();

    for (const ln of lines) {
      const lineDiv = document.createElement('div');
      lineDiv.className = 'my-rules__mirror-line';
      lineDiv.textContent = ln.length > 0 ? ln : ' ';
      mirrorLines.appendChild(lineDiv);
    }

    const newWraps: number[] = [];
    const children = mirrorLines.children;
    for (let i = 0; i < children.length; i++) {
      const el = children[i] as HTMLElement;
      const h = el.getBoundingClientRect().height;
      const ratio = h / lineHeightPx;

      const wrapsForLine = Math.max(1, Math.floor(ratio + 0.001));

      newWraps.push(wrapsForLine);
    }

    setWraps(newWraps);
  }, [computeLineHeightPx, lines, textAreaRef, mirrorRef]);

  useLayoutEffect(() => {
    recalcWraps();
  }, [value, lines, recalcWraps]);

  useEffect(() => {
    const ta = textAreaRef.current;
    if (!ta) return;
    const ro = new ResizeObserver(() => recalcWraps());
    ro.observe(ta);
    return () => ro.disconnect();
  }, [recalcWraps, textAreaRef]);

  return { wraps, recalcWraps };
}
