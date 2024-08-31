import { useLayoutEffect, useRef } from 'react';
import type { RefObject } from 'react';

type Callback = (entry: ResizeObserverEntry) => void;

const keyToElementsMap = new Map<string, Element[]>();
const keyToCallbacksMap = new Map<string, Callback[]>();
const elementToKeyMap = new WeakMap<Element, string>();

let globalObserver: ResizeObserver | null = null;

function getGlobalObserver(): ResizeObserver {
  if (!globalObserver) {
    globalObserver = new ResizeObserver((entries) => {
      entries.forEach((entry) => {
        const element = entry.target;
        const key = elementToKeyMap.get(element);
        if (key) {
          const callbacks = keyToCallbacksMap.get(key);
          callbacks?.forEach((callback) => callback(entry));
        }
      });
    });
  }

  return globalObserver;
}

function observe(element: Element, callback: Callback, key: string): void {
  const observer = getGlobalObserver();
  let elements = keyToElementsMap.get(key);
  if (!elements) {
    elements = [];
    keyToElementsMap.set(key, elements);
  }
  elements.push(element);
  elementToKeyMap.set(element, key);

  let callbacks = keyToCallbacksMap.get(key);
  if (!callbacks) {
    callbacks = [];
    keyToCallbacksMap.set(key, callbacks);
    observer.observe(element);
  }
  callbacks.push(callback);
}

function unobserve(element: Element, callback: Callback, key: string): void {
  const callbacks = keyToCallbacksMap.get(key);
  if (callbacks) {
    const index = callbacks.indexOf(callback);
    if (index !== -1) {
      callbacks.splice(index, 1);
    }
    if (callbacks.length === 0) {
      keyToCallbacksMap.delete(key);
    }
  }

  const elements = keyToElementsMap.get(key);
  if (elements) {
    const idx = elements.indexOf(element);
    if (idx !== -1) {
      elements.splice(idx, 1);
      elementToKeyMap.delete(element);
      if (idx === 0 && elements.length > 0) {
        globalObserver?.unobserve(element);
        globalObserver?.observe(elements[0]);
      }
    }
    if (elements.length === 0) {
      keyToElementsMap.delete(key);
      globalObserver?.unobserve(element);
    }
  }
}

export function useResizeObserver(
  elementRef: RefObject<Element>,
  callback: Callback,
  key?: string,
): void {
  const storedKey = useRef(
    key || `key_${Math.random().toString(36).slice(2, 9)}`,
  ).current;

  useLayoutEffect(() => {
    const element = elementRef.current;
    if (element) {
      observe(element, callback, storedKey);

      return () => {
        if (element) {
          unobserve(element, callback, storedKey);
        }
      };
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);
}
