import { parse } from './parse';
import { QueryRunner } from './queryExec';

export default function (rules: string): void {
  const lines = rules
    .split('\n')
    .map((r) => r.trim())
    .filter((l) => l.length > 0);
  const runners: QueryRunner[] = [];

  for (const l of lines) {
    try {
      const query = parse(l);
      const runner = new QueryRunner(query);
      runners.push(runner);
    } catch (ex) {
      console.debug(`Failed to parse line ${l}: ${ex}`);
    }
  }

  const runAll = () => {
    for (const r of runners) {
      try {
        r.run();
      } catch (ex) {
        console.debug(`Failed to run ${r}: ${ex}`);
      }
    }
  };

  document.addEventListener('DOMContentLoaded', () => {
    runAll();
  });

  window.addEventListener('popstate', () => {
    runAll();
  });
}
