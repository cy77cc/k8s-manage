import { useEffect } from 'react';
import type { DependencyList } from 'react';

export const useVisibilityRefresh = (
  callback: () => void,
  intervalMs: number,
  deps: DependencyList = [],
) => {
  useEffect(() => {
    if (intervalMs <= 0) return;

    let timer: ReturnType<typeof setInterval> | null = null;

    const tick = () => {
      if (typeof document === 'undefined' || document.visibilityState === 'visible') {
        callback();
      }
    };

    timer = setInterval(tick, intervalMs);

    const onVisible = () => {
      if (document.visibilityState === 'visible') {
        callback();
      }
    };

    document.addEventListener('visibilitychange', onVisible);
    return () => {
      if (timer) clearInterval(timer);
      document.removeEventListener('visibilitychange', onVisible);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [callback, intervalMs, ...deps]);
};
