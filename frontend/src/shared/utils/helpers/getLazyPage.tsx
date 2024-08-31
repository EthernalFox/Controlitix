import type { ComponentType } from 'react';

interface Module {
  default: ComponentType
}

export const getLazyPage = (importCallback: () => Promise<Module>) => async () => {
  const page = await importCallback();

  return { Component: page.default };
};
