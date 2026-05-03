import type { CSSProperties } from "react";

import styles from "./AmbientOrbs.module.css";

type AmbientOrbsProps = {
  animated?: boolean;
};

type OrbStyle = CSSProperties & {
  "--orb": string;
};

const getOrbClassName = (animated: boolean) =>
  animated ? `${styles.orb} ${styles.animated}` : styles.orb;

export const AmbientOrbs = ({ animated = false }: AmbientOrbsProps) => {
  return (
    <div className={styles.root} aria-hidden>
      <span
        className={getOrbClassName(animated)}
        style={{ "--orb": "var(--ctrx-orb-1)", left: "10%", top: "10%" } as OrbStyle}
      />
      <span
        className={getOrbClassName(animated)}
        style={{ "--orb": "var(--ctrx-orb-2)", right: "15%", top: "20%" } as OrbStyle}
      />
      <span
        className={getOrbClassName(animated)}
        style={{ "--orb": "var(--ctrx-orb-3)", left: "20%", bottom: "10%" } as OrbStyle}
      />
      <span
        className={getOrbClassName(animated)}
        style={{ "--orb": "var(--ctrx-orb-4)", right: "10%", bottom: "15%" } as OrbStyle}
      />
    </div>
  );
};
