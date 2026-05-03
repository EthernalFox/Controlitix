import type { SVGProps } from "react";

const baseProps: SVGProps<SVGSVGElement> = {
  viewBox: "0 0 8 8",
  fill: "none",
  width: 8,
  height: 8,
  xmlns: "http://www.w3.org/2000/svg"
};

export const CircleIcon = () => (
  <svg {...baseProps}>
    <circle cx="4" cy="4" r="3.2" fill="currentColor" />
  </svg>
);

export const TriangleIcon = () => (
  <svg {...baseProps}>
    <path d="M4 0.8L7.2 7H0.8L4 0.8Z" fill="currentColor" />
  </svg>
);

export const DiamondIcon = () => (
  <svg {...baseProps}>
    <path d="M4 0.8L7.2 4L4 7.2L0.8 4L4 0.8Z" fill="currentColor" />
  </svg>
);

export const RingHollowIcon = () => (
  <svg {...baseProps}>
    <circle cx="4" cy="4" r="3" stroke="currentColor" strokeWidth="1.4" />
  </svg>
);

export const HatchedIcon = () => (
  <svg {...baseProps}>
    <rect x="1" y="1" width="6" height="6" rx="1.2" stroke="currentColor" strokeWidth="1" />
    <path d="M1.4 5.8L5.8 1.4" stroke="currentColor" strokeWidth="1" />
    <path d="M2.4 6.8L6.8 2.4" stroke="currentColor" strokeWidth="1" />
  </svg>
);

export const RingBrokenIcon = () => (
  <svg {...baseProps}>
    <path d="M6.2 2.2A3 3 0 1 0 5.8 6.2" stroke="currentColor" strokeWidth="1.2" />
    <path d="M4.8 1.1L6.9 1.1L6.9 3.2" stroke="currentColor" strokeWidth="1.2" />
  </svg>
);

export const SquareDimmedIcon = () => (
  <svg {...baseProps}>
    <rect x="1" y="1" width="6" height="6" rx="1.2" fill="currentColor" opacity="0.7" />
  </svg>
);

export const CheckCircleIcon = () => (
  <svg {...baseProps}>
    <circle cx="4" cy="4" r="3" stroke="currentColor" strokeWidth="1.2" />
    <path d="M2.4 4.1L3.6 5.3L5.8 3" stroke="currentColor" strokeWidth="1.2" />
  </svg>
);

