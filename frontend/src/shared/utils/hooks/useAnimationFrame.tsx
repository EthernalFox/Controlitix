import { useEffect, useRef } from 'react';

const useAnimationFrame = (callback: (timestamp: number) => void, deps: unknown[]) => {
  const requestRef = useRef<number>();

  const animate = (timestamp: number) => {
    callback(timestamp);
    requestRef.current = requestAnimationFrame(animate);
  };

  useEffect(() => {
    requestRef.current = requestAnimationFrame(animate);

    return () => cancelAnimationFrame(requestRef?.current?? 0);
    
  }, (deps));
};

export default useAnimationFrame;
