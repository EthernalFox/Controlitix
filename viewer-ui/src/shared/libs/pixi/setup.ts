import { extend } from "@pixi/react";
import { Assets, Container, Graphics, Sprite, Text, Texture } from "pixi.js";

let pixiExtended = false;

const texturePromises = new Map<string, Promise<Texture | null>>();

export const ensurePixiSetup = (): void => {
  if (pixiExtended) {
    return;
  }

  extend({
    Container,
    Graphics,
    Sprite,
    Text
  });

  pixiExtended = true;
};

export const loadTexture = async (source: string): Promise<Texture | null> => {
  const normalizedSource = source.trim();
  if (!normalizedSource) {
    return null;
  }

  const existing = texturePromises.get(normalizedSource);
  if (existing) {
    return existing;
  }

  const loadingPromise = (async () => {
    try {
      const texture = await Assets.load<Texture>(normalizedSource);
      return texture;
    } catch (error) {
      console.error(`[diagram] failed to load image texture: ${normalizedSource}`, error);
      return null;
    }
  })();

  texturePromises.set(normalizedSource, loadingPromise);

  return loadingPromise;
};

export const unloadTextures = async (sources: string[]): Promise<void> => {
  const uniqueSources = Array.from(
    new Set(
      sources
        .map((source) => source.trim())
        .filter(Boolean)
    )
  );

  await Promise.all(
    uniqueSources.map(async (source) => {
      try {
        await Assets.unload(source);
      } catch {
        // Texture may be already unloaded; noop.
      }

      texturePromises.delete(source);
    })
  );
};
