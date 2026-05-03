import { Outlet } from "react-router";

import { Header, Layout, Logo, Navbar } from "@shared/ui";
import { ObjectsNavbar } from "@widgets/ObjectsNavbar";
import { PageTitle } from "@widgets/PageTitle";
import { UserMenu } from "@widgets/UserMenu";

const ListLayout = () => {
  return (
    <Layout
      mode="list"
      header={<Header before={<Logo />} main={<PageTitle />} after={<UserMenu />} />}
      navbar={<Navbar center={<ObjectsNavbar />} />}
    >
      <Outlet />
    </Layout>
  );
};

export default ListLayout;
