import { Outlet } from "react-router-dom";
import type { ElementType } from "react";


const Providers = ({ providers }: {providers: ElementType[]}) => (
  <>
    {providers.reduceRight(
      (acc, Component) => (
        <Component>{acc}</Component>
      ),
      <Outlet />
    )}
  </>
);
export default Providers;
